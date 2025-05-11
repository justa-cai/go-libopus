package opus

import (
	"errors"
	"runtime"
	"sync"
	"unsafe"

	"github.com/ebitengine/purego"
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

	libopus uintptr

	opus_encoder_create  func(uint32, uint32, int32, uintptr) uintptr
	opus_encoder_destroy func(uintptr)
	opus_encode          func(uintptr, uintptr, int32, uintptr, int32) int32
	opus_decoder_create  func(uint32, uint32, uintptr) uintptr
	opus_decoder_destroy func(uintptr)
	opus_decode          func(uintptr, uintptr, int32, uintptr, int32) int32
	opus_strerror        func(int32) uintptr
)

func loadOpus() error {
	opusOnce.Do(func() {
		var libName string
		switch runtime.GOOS {
		case "windows":
			libName = "opus.dll"
		case "darwin":
			libName = "libopus.dylib"
		default:
			libName = "libopus.so"
		}
		lib, err := purego.Dlopen(libName, purego.RTLD_LAZY|purego.RTLD_LOCAL)
		if err != nil {
			opusErr = err
			return
		}
		libopus = lib
		purego.RegisterLibFunc(&opus_encoder_create, libopus, "opus_encoder_create")
		purego.RegisterLibFunc(&opus_encoder_destroy, libopus, "opus_encoder_destroy")
		purego.RegisterLibFunc(&opus_encode, libopus, "opus_encode")
		purego.RegisterLibFunc(&opus_decoder_create, libopus, "opus_decoder_create")
		purego.RegisterLibFunc(&opus_decoder_destroy, libopus, "opus_decoder_destroy")
		purego.RegisterLibFunc(&opus_decode, libopus, "opus_decode")
		purego.RegisterLibFunc(&opus_strerror, libopus, "opus_strerror")
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
	encoder := opus_encoder_create(uint32(sampleRate), uint32(channels), int32(application), uintptr(unsafe.Pointer(&errCode)))
	if errCode != 0 {
		return nil, errors.New(opusStrError(errCode))
	}
	return &OpusEncoder{encoder: encoder}, nil
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
	decoder := opus_decoder_create(uint32(sampleRate), uint32(channels), uintptr(unsafe.Pointer(&errCode)))
	if errCode != 0 {
		return nil, errors.New(opusStrError(errCode))
	}
	return &OpusDecoder{decoder: decoder}, nil
}

// Encode 编码音频数据
func (e *OpusEncoder) Encode(input []byte, output []byte) (int, error) {
	if e.encoder == 0 {
		return 0, errors.New("encoder not initialized")
	}
	inPtr := unsafe.Pointer(&input[0])
	outPtr := unsafe.Pointer(&output[0])
	frameSize := len(input) / 2 // int16
	ret := opus_encode(e.encoder, uintptr(inPtr), int32(frameSize), uintptr(outPtr), int32(len(output)))
	if ret < 0 {
		return int(ret), errors.New(opusStrError(ret))
	}
	return int(ret), nil
}

// Decode 解码音频数据
func (d *OpusDecoder) Decode(input []byte, output []byte) (int, error) {
	if d.decoder == 0 {
		return 0, errors.New("decoder not initialized")
	}
	inPtr := unsafe.Pointer(&input[0])
	outPtr := unsafe.Pointer(&output[0])
	ret := opus_decode(d.decoder, uintptr(inPtr), int32(len(input)), uintptr(outPtr), int32(len(output)/2))
	if ret < 0 {
		return int(ret), errors.New(opusStrError(ret))
	}
	return int(ret), nil
}

// Close 释放编码器资源
func (e *OpusEncoder) Close() {
	if e.encoder != 0 {
		opus_encoder_destroy(e.encoder)
		e.encoder = 0
	}
}

// Close 释放解码器资源
func (d *OpusDecoder) Close() {
	if d.decoder != 0 {
		opus_decoder_destroy(d.decoder)
		d.decoder = 0
	}
}

// opusStrError 获取错误字符串
func opusStrError(code int32) string {
	ret := opus_strerror(code)
	return unsafeString(ret)
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
