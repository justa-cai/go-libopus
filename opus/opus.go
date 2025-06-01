package opus

/*
#cgo CFLAGS: -I${SRCDIR}/include/opus
#cgo LDFLAGS: -L${SRCDIR} -lopus
#include <opus.h>
static int go_opus_encoder_set_bitrate(OpusEncoder *enc, opus_int32 bitrate) {
    return opus_encoder_ctl(enc, OPUS_SET_BITRATE(bitrate));
}
static int go_opus_encoder_set_complexity(OpusEncoder *enc, int complexity) {
    return opus_encoder_ctl(enc, OPUS_SET_COMPLEXITY(complexity));
}
static int go_opus_encoder_set_signal(OpusEncoder *enc, int signal) {
    return opus_encoder_ctl(enc, OPUS_SET_SIGNAL(signal));
}
*/
import "C"
import (
	"errors"
	"unsafe"
)

// OpusApplication constants
const (
	OpusApplicationVoIP     = 2048
	OpusApplicationAudio    = 2049
	OpusApplicationLowDelay = 2051
)

// Opus encoder control constants
const (
	OPUS_SET_BITRATE_REQUEST     = 4002
	OPUS_SET_COMPLEXITY_REQUEST  = 4010
	OPUS_SET_SIGNAL_REQUEST      = 4024
	OPUS_SET_APPLICATION_REQUEST = 4000
)

// Opus signal types
const (
	OPUS_SIGNAL_AUTO  = -1000
	OPUS_SIGNAL_VOICE = 3001
	OPUS_SIGNAL_MUSIC = 3002
)

// OpusEncoder represents an Opus encoder
type OpusEncoder struct {
	encoder *C.OpusEncoder
}

// OpusDecoder represents an Opus decoder
type OpusDecoder struct {
	decoder *C.OpusDecoder
}

// NewEncoder creates a new Opus encoder
func NewEncoder(sampleRate int, channels int, application int) (*OpusEncoder, error) {
	if sampleRate <= 0 || channels <= 0 || application < 0 {
		return nil, errors.New("invalid parameter: must be positive")
	}

	var err C.int
	encoder := C.opus_encoder_create(C.opus_int32(sampleRate), C.int(channels), C.int(application), &err)
	if err != 0 {
		return nil, errors.New(C.GoString(C.opus_strerror(err)))
	}

	return &OpusEncoder{encoder: encoder}, nil
}

// SetBitrate sets the bitrate for the encoder
func (e *OpusEncoder) SetBitrate(bitrate int) error {
	if e.encoder == nil {
		return errors.New("encoder not initialized")
	}
	ret := C.go_opus_encoder_set_bitrate(e.encoder, C.opus_int32(bitrate))
	if ret != 0 {
		return errors.New(C.GoString(C.opus_strerror(ret)))
	}
	return nil
}

// SetComplexity sets the complexity for the encoder
func (e *OpusEncoder) SetComplexity(complexity int) error {
	if e.encoder == nil {
		return errors.New("encoder not initialized")
	}
	ret := C.go_opus_encoder_set_complexity(e.encoder, C.int(complexity))
	if ret != 0 {
		return errors.New(C.GoString(C.opus_strerror(ret)))
	}
	return nil
}

// SetSignal sets the signal type for the encoder
func (e *OpusEncoder) SetSignal(signal int) error {
	if e.encoder == nil {
		return errors.New("encoder not initialized")
	}
	ret := C.go_opus_encoder_set_signal(e.encoder, C.int(signal))
	if ret != 0 {
		return errors.New(C.GoString(C.opus_strerror(ret)))
	}
	return nil
}

// NewDecoder creates a new Opus decoder
func NewDecoder(sampleRate int, channels int) (*OpusDecoder, error) {
	if sampleRate <= 0 || channels <= 0 {
		return nil, errors.New("invalid parameter: must be positive")
	}

	var err C.int
	decoder := C.opus_decoder_create(C.opus_int32(sampleRate), C.int(channels), &err)
	if err != 0 {
		return nil, errors.New(C.GoString(C.opus_strerror(err)))
	}

	return &OpusDecoder{decoder: decoder}, nil
}

// Encode encodes audio data
func (e *OpusEncoder) Encode(input []byte, output []byte) (int, error) {
	if e.encoder == nil {
		return 0, errors.New("encoder not initialized")
	}
	if len(input) == 0 {
		return 0, errors.New("empty input")
	}
	if len(output) == 0 {
		return 0, errors.New("empty output buffer")
	}

	pcm := (*C.opus_int16)(unsafe.Pointer(&input[0]))
	data := (*C.uchar)(unsafe.Pointer(&output[0]))
	frameSize := len(input) / 2 // int16, 单通道

	ret := C.opus_encode(
		e.encoder,
		pcm,
		C.int(frameSize),
		data,
		C.opus_int32(len(output)),
	)

	if ret < 0 {
		return int(ret), errors.New(C.GoString(C.opus_strerror(C.int(ret))))
	}

	return int(ret), nil
}

// Decode decodes audio data
func (d *OpusDecoder) Decode(input []byte, output []byte) (int, error) {
	if d.decoder == nil {
		return 0, errors.New("decoder not initialized")
	}
	if len(input) == 0 {
		return 0, errors.New("empty input")
	}
	if len(output) == 0 {
		return 0, errors.New("empty output buffer")
	}

	data := (*C.uchar)(unsafe.Pointer(&input[0]))
	pcm := (*C.opus_int16)(unsafe.Pointer(&output[0]))

	ret := C.opus_decode(
		d.decoder,
		data,
		C.opus_int32(len(input)),
		pcm,
		C.int(len(output)/2), // 单通道，2字节每采样
		0,                    // decode_fec
	)

	if ret < 0 {
		return int(ret), errors.New(C.GoString(C.opus_strerror(C.int(ret))))
	}

	return int(ret), nil
}

// Close frees the encoder resources
func (e *OpusEncoder) Close() {
	if e.encoder != nil {
		C.opus_encoder_destroy(e.encoder)
		e.encoder = nil
	}
}

// Close frees the decoder resources
func (d *OpusDecoder) Close() {
	if d.decoder != nil {
		C.opus_decoder_destroy(d.decoder)
		d.decoder = nil
	}
}
