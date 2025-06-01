package opus_test

import (
	"testing"

	"github.com/justa-cai/go-libopus/opus"
)

func TestOpusEncoder(t *testing.T) {
	// 测试创建编码器
	encoder, err := opus.NewEncoder(48000, 2, opus.OpusApplicationAudio)
	if err != nil {
		t.Fatalf("Failed to create encoder: %v", err)
	}
	defer encoder.Close()

	// 测试无效参数
	_, err = opus.NewEncoder(0, 2, opus.OpusApplicationAudio)
	if err == nil {
		t.Error("Expected error for invalid sample rate")
	}

	_, err = opus.NewEncoder(48000, 0, opus.OpusApplicationAudio)
	if err == nil {
		t.Error("Expected error for invalid channels")
	}

	_, err = opus.NewEncoder(48000, 2, -1)
	if err == nil {
		t.Error("Expected error for invalid application")
	}
}

func TestOpusDecoder(t *testing.T) {
	// 测试创建解码器
	decoder, err := opus.NewDecoder(48000, 2)
	if err != nil {
		t.Fatalf("Failed to create decoder: %v", err)
	}
	defer decoder.Close()

	// 测试无效参数
	_, err = opus.NewDecoder(0, 2)
	if err == nil {
		t.Error("Expected error for invalid sample rate")
	}

	_, err = opus.NewDecoder(48000, 0)
	if err == nil {
		t.Error("Expected error for invalid channels")
	}
}

func TestOpusEncodeDecode(t *testing.T) {
	// 创建测试数据 (10ms of 48kHz stereo audio)
	frameSize := 480                     // 10ms at 48kHz
	input := make([]byte, frameSize*2*2) // 2 channels, 2 bytes per sample

	// 创建编码器
	encoder, err := opus.NewEncoder(48000, 2, opus.OpusApplicationAudio)
	if err != nil {
		t.Fatalf("Failed to create encoder: %v", err)
	}
	defer encoder.Close()

	// 编码
	output := make([]byte, 1024) // 足够大的缓冲区
	nBytes, err := encoder.Encode(input, output)
	if err != nil {
		t.Fatalf("Failed to encode: %v", err)
	}
	if nBytes <= 0 {
		t.Error("Expected positive number of encoded bytes")
	}

	// 创建解码器
	decoder, err := opus.NewDecoder(48000, 2)
	if err != nil {
		t.Fatalf("Failed to create decoder: %v", err)
	}
	defer decoder.Close()

	// 解码
	decoded := make([]byte, frameSize*2*2)
	nSamples, err := decoder.Decode(output[:nBytes], decoded)
	if err != nil {
		t.Fatalf("Failed to decode: %v", err)
	}
	if nSamples != frameSize {
		t.Errorf("Expected %d samples, got %d", frameSize, nSamples)
	}
}

func TestOpusErrorHandling(t *testing.T) {
	// 测试空编码器
	encoder := &opus.OpusEncoder{}
	_, err := encoder.Encode([]byte{1, 2, 3}, make([]byte, 10))
	if err == nil {
		t.Error("Expected error for uninitialized encoder")
	}

	// 测试空解码器
	decoder := &opus.OpusDecoder{}
	_, err = decoder.Decode([]byte{1, 2, 3}, make([]byte, 10))
	if err == nil {
		t.Error("Expected error for uninitialized decoder")
	}

	// 测试无效的输入数据
	encoder, err = opus.NewEncoder(48000, 2, opus.OpusApplicationAudio)
	if err != nil {
		t.Fatalf("Failed to create encoder: %v", err)
	}
	defer encoder.Close()

	_, err = encoder.Encode([]byte{}, make([]byte, 10))
	if err == nil {
		t.Error("Expected error for empty input")
	}

	decoder, err = opus.NewDecoder(48000, 2)
	if err != nil {
		t.Fatalf("Failed to create decoder: %v", err)
	}
	defer decoder.Close()

	_, err = decoder.Decode([]byte{}, make([]byte, 10))
	if err == nil {
		t.Error("Expected error for empty input")
	}
}
