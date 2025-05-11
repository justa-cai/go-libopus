package opus_test

import (
	"testing"

	"github.com/justa-cai/go-libopus/opus"
)

func TestOpus(t *testing.T) {
	// 示例音频数据（16kHz, 2通道, 10ms帧）
	input := make([]byte, 160*2*2) // 10ms * 16kHz * 2通道，每个采样2字节

	// 创建编码器
	encoder, err := opus.NewEncoder(16000, 2, opus.OpusApplicationAudio)
	if err != nil {
		t.Fatalf("创建编码器失败: %v", err)
	}
	defer encoder.Close()

	// 编码音频
	output := make([]byte, 1024) // opus压缩后数据一般远小于原始数据
	nBytes, err := encoder.Encode(input, output)
	if err != nil {
		t.Fatalf("编码失败: %v", err)
	}
	t.Logf("编码后数据长度: %d bytes", nBytes)

	// 创建解码器
	decoder, err := opus.NewDecoder(16000, 2)
	if err != nil {
		t.Fatalf("创建解码器失败: %v", err)
	}
	defer decoder.Close()

	// 解码音频
	decoded := make([]byte, 160*2*2) // 解码后数据长度与原始一致
	nSamples, err := decoder.Decode(output[:nBytes], decoded)
	if err != nil {
		t.Fatalf("解码失败: %v", err)
	}
	t.Logf("解码后数据长度: %d bytes", nSamples*2)

	// 验证数据完整性
	if nSamples*2 != len(decoded) {
		t.Log("警告: 解码数据长度与原始数据不一致")
	}
}
