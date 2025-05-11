# go-libopus

## 简介
go-libopus 是一个基于 libopus 的 Golang 跨平台库，基于DLOPEN技术，方便解决不同平台上的CGO的编译，以及链接问题。

## 支持平台

- Linux
- macOS
- Windows

## 安装

```bash
go get github.com/justa-cai/go-libopus
```

## 依赖

- Go 1.18 及以上（建议使用最新版）
- [libopus](https://opus-codec.org/) 开发库  
  安装方式：
  - **Debian/Ubuntu**：
    ```bash
    sudo apt-get install libopus-dev
    ```
  - **macOS**（使用 Homebrew）：
    ```bash
    brew install opus
    ```
  - **Windows**：
    1. 推荐使用 [MSYS2](https://www.msys2.org/) 环境，安装命令：
       ```bash
       pacman -S mingw-w64-x86_64-opus
       ```
    2. 或者从 [opus 官网](https://opus-codec.org/downloads/) 下载预编译的 DLL 并配置到 PATH。

## 快速开始

```go
package main

import (
	"fmt"
	"github.com/justa-cai/go-libopus/opus"
)

func main() {
	// 创建编码器
	encoder, err := opus.NewEncoder(16000, 2, opus.OpusApplicationAudio)
	if err != nil {
		panic(err)
	}
	defer encoder.Close()

	// 编码
	input := make([]byte, 160*2*2) // 10ms, 16kHz, 2通道
	output := make([]byte, 1024)
	n, err := encoder.Encode(input, output)
	if err != nil {
		panic(err)
	}
	fmt.Println("编码后字节数:", n)

	// 创建解码器
	decoder, err := opus.NewDecoder(16000, 2)
	if err != nil {
		panic(err)
	}
	defer decoder.Close()

	// 解码
	decoded := make([]byte, 160*2*2)
	nSamples, err := decoder.Decode(output[:n], decoded)
	if err != nil {
		panic(err)
	}
	fmt.Println("解码后采样数:", nSamples)
}
```

## 主要 API

- `opus.NewEncoder(sampleRate int, channels int, application int) (*OpusEncoder, error)`  
  创建 Opus 编码器

- `opus.NewDecoder(sampleRate int, channels int) (*OpusDecoder, error)`  
  创建 Opus 解码器

- `(*OpusEncoder) Encode(input []byte, output []byte) (int, error)`  
  编码 PCM 数据为 Opus

- `(*OpusDecoder) Decode(input []byte, output []byte) (int, error)`  
  解码 Opus 数据为 PCM

- `(*OpusEncoder) Close()` / `(*OpusDecoder) Close()`  
  释放资源

## 构建

```bash
git clone https://github.com/justa-cai/go-libopus.git
cd go-libopus
go build ./...
```

## 贡献

欢迎提交 Issue 和 PR！

## 许可证

本项目采用 MIT 许可证，详见 `LICENSE` 文件。

## Star 趋势

[![Star History Chart](https://api.star-history.com/svg?repos=justa-cai/go-libopus&type=Date)](https://star-history.com/#justa-cai/go-libopus)

