package opus

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"syscall"
	"unsafe"
)

// OpusEncoder 封装Opus编码器
// 这里用uintptr存储C对象指针
type OpusEncoder struct {
	encoder uintptr
}

type OpusDecoder struct {
	decoder uintptr
}

// OpusApplication constants
const (
	OpusApplicationVoIP     = 2048
	OpusApplicationAudio    = 2049
	OpusApplicationLowDelay = 2051
)

var (
	opusOnce sync.Once
	opusErr  error

	libopus *syscall.DLL

	opus_encoder_create  *syscall.Proc
	opus_encoder_destroy *syscall.Proc
	opus_encode          *syscall.Proc
	opus_decoder_create  *syscall.Proc
	opus_decoder_destroy *syscall.Proc
	opus_decode          *syscall.Proc
	opus_strerror        *syscall.Proc
)

func loadOpus() error {
	opusOnce.Do(func() {
		var libName string

		switch runtime.GOOS {
		case "windows":
			// Check if opus.dll exists in x86 or x64 directory based on architecture
			currentDir, err := os.Getwd()
			if err != nil {
				opusErr = fmt.Errorf("failed to get current directory: %v", err)
				return
			}

			// Determine architecture-specific directory
			archDir := "x64"
			if runtime.GOARCH == "386" {
				archDir = "x86"
			}

			// Try to load from architecture-specific directory first
			archDllPath := filepath.Join(currentDir, archDir, "opus.dll")
			if _, err := os.Stat(archDllPath); err == nil {
				libName = archDllPath
			} else {
				// Fallback to current directory
				localDllPath := filepath.Join(currentDir, "opus.dll")
				if _, err := os.Stat(localDllPath); err == nil {
					libName = localDllPath
				} else {
					libName = "opus.dll"
				}
			}
			fmt.Println("libName:", libName)
		case "darwin":
			libName = "libopus.dylib"
		default:
			libName = "libopus.so"
		}

		lib, err := syscall.LoadDLL(libName)
		if err != nil {
			opusErr = fmt.Errorf("failed to load %s: %v", libName, err)
			return
		}
		libopus = lib

		// Load functions using FindProc
		opus_encoder_create = lib.MustFindProc("opus_encoder_create")
		opus_encoder_destroy = lib.MustFindProc("opus_encoder_destroy")
		opus_encode = lib.MustFindProc("opus_encode")
		opus_decoder_create = lib.MustFindProc("opus_decoder_create")
		opus_decoder_destroy = lib.MustFindProc("opus_decoder_destroy")
		opus_decode = lib.MustFindProc("opus_decode")
		opus_strerror = lib.MustFindProc("opus_strerror")
	})
	return opusErr
}

// NewEncoder 创建并初始化Opus编码器
func NewEncoder(sampleRate int, channels int, application int) (*OpusEncoder, error) {
	if err := loadOpus(); err != nil {
		return nil, err
	}
	if sampleRate <= 0 || channels <= 0 || application < 0 {
		return nil, errors.New("invalid parameter: must be positive")
	}
	var errCode int32
	ret, _, _ := opus_encoder_create.Call(
		uintptr(sampleRate),
		uintptr(channels),
		uintptr(application),
		uintptr(unsafe.Pointer(&errCode)),
	)
	if errCode != 0 {
		return nil, errors.New(opusStrError(errCode))
	}
	return &OpusEncoder{encoder: ret}, nil
}

// NewDecoder 创建并初始化Opus解码器
func NewDecoder(sampleRate int, channels int) (*OpusDecoder, error) {
	if err := loadOpus(); err != nil {
		return nil, err
	}
	if sampleRate <= 0 || channels <= 0 {
		return nil, errors.New("invalid parameter: must be positive")
	}
	var errCode int32
	ret, _, _ := opus_decoder_create.Call(
		uintptr(sampleRate),
		uintptr(channels),
		uintptr(unsafe.Pointer(&errCode)),
	)
	if errCode != 0 {
		return nil, errors.New(opusStrError(errCode))
	}
	return &OpusDecoder{decoder: ret}, nil
}

// Encode 编码音频数据
func (e *OpusEncoder) Encode(input []byte, output []byte) (int, error) {
	if e.encoder == 0 {
		return 0, errors.New("encoder not initialized")
	}
	pcm := (*int16)(unsafe.Pointer(&input[0]))
	data := (*byte)(unsafe.Pointer(&output[0]))
	frameSize := len(input) / 2 // int16
	ret, _, _ := opus_encode.Call(
		e.encoder,
		uintptr(unsafe.Pointer(pcm)),
		uintptr(frameSize),
		uintptr(unsafe.Pointer(data)),
		uintptr(len(output)),
	)
	if ret < 0 {
		return int(ret), errors.New(opusStrError(int32(ret)))
	}
	return int(ret), nil
}

// Decode 解码音频数据
func (d *OpusDecoder) Decode(input []byte, output []byte) (int, error) {
	if d.decoder == 0 {
		return 0, errors.New("decoder not initialized")
	}
	data := (*byte)(unsafe.Pointer(&input[0]))
	pcm := (*int16)(unsafe.Pointer(&output[0]))
	ret, _, _ := opus_decode.Call(
		d.decoder,
		uintptr(unsafe.Pointer(data)),
		uintptr(len(input)),
		uintptr(unsafe.Pointer(pcm)),
		uintptr(len(output)/2),
	)
	if ret < 0 {
		return int(ret), errors.New(opusStrError(int32(ret)))
	}
	return int(ret), nil
}

// Close 释放编码器资源
func (e *OpusEncoder) Close() {
	if e.encoder != 0 {
		opus_encoder_destroy.Call(e.encoder)
		e.encoder = 0
	}
}

// Close 释放解码器资源
func (d *OpusDecoder) Close() {
	if d.decoder != 0 {
		opus_decoder_destroy.Call(d.decoder)
		d.decoder = 0
	}
}

// opusStrError 获取错误字符串
func opusStrError(code int32) string {
	ret, _, _ := opus_strerror.Call(uintptr(code))
	if ret == 0 {
		return ""
	}
	return unsafe.String((*byte)(unsafe.Pointer(ret)), 1024) // 假设错误消息不会超过1024字节
}

// unsafeString 将 C 字符串指针转为 Go 字符串
func unsafeString(ptr uintptr) string {
	if ptr == 0 {
		return ""
	}
	cstr := (*[1 << 20]byte)(unsafe.Pointer(ptr))
	for i := 0; i < len(cstr); i++ {
		if cstr[i] == 0 {
			return string(cstr[:i])
		}
	}
	return string(cstr[:])
}
