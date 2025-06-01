// Package ogg provides Go bindings for libogg
package ogg

// #cgo CFLAGS: -I${SRCDIR}/include
// #cgo LDFLAGS: -L${SRCDIR}/ -logg
// #include <ogg/ogg.h>
// #include <stdlib.h>
// #include <string.h>
// extern int ogg_stream_pagein(ogg_stream_state *os, ogg_page *og);
// extern int ogg_stream_packetout(ogg_stream_state *os, ogg_packet *op);
import "C"
import (
	"errors"
	"unsafe"
)

// OggSyncState represents the ogg_sync_state structure from libogg
// It is used for synchronizing Ogg bitstreams
type OggSyncState struct {
	state C.ogg_sync_state
}

// OggStreamState represents the ogg_stream_state structure from libogg
// It is used for managing Ogg streams
type OggStreamState struct {
	state C.ogg_stream_state
}

// OggPacket represents the ogg_packet structure from libogg
// It contains a single Ogg packet with its metadata
type OggPacket struct {
	Packet     []byte // The packet data
	Bytes      int    // Number of bytes in the packet
	BOS        int    // Beginning of stream flag
	EOS        int    // End of stream flag
	Granulepos int64  // Position in the stream
	Packetno   int64  // Packet sequence number
}

// OggPage represents the ogg_page structure from libogg
// It contains a single Ogg page with its header and body
type OggPage struct {
	Header    []byte // Page header
	HeaderLen int    // Length of the header
	Body      []byte // Page body
	BodyLen   int    // Length of the body
}

// NewOggSyncState 初始化Ogg同步状态
func NewOggSyncState() (*OggSyncState, error) {
	state := &OggSyncState{}
	ret := C.ogg_sync_init(&state.state)
	if ret != 0 {
		return nil, errors.New("failed to initialize ogg sync state")
	}
	return state, nil
}

// Clear 清理Ogg同步状态
func (s *OggSyncState) Clear() error {
	if ret := C.ogg_sync_clear(&s.state); ret != 0 {
		return errors.New("failed to clear ogg sync state")
	}
	return nil
}

// Buffer 获取同步缓冲区
func (s *OggSyncState) Buffer(size int) ([]byte, error) {
	buffer := C.ogg_sync_buffer(&s.state, C.long(size))
	if buffer == nil {
		return nil, errors.New("failed to get buffer")
	}
	return unsafe.Slice((*byte)(unsafe.Pointer(buffer)), size), nil
}

// Wrote 标记已写入的字节数
func (s *OggSyncState) Wrote(bytes int) error {
	if ret := C.ogg_sync_wrote(&s.state, C.long(bytes)); ret != 0 {
		return errors.New("failed to mark written bytes")
	}
	return nil
}

// PageOut 从同步状态中提取页面
func (s *OggSyncState) PageOut(page *OggPage) (int, error) {
	var cPage C.ogg_page
	ret := C.ogg_sync_pageout(&s.state, &cPage)
	if ret < 0 {
		return 0, errors.New("failed to get page from sync state")
	}

	// 转换C结构体到Go结构体
	page.Header = unsafe.Slice((*byte)(unsafe.Pointer(cPage.header)), cPage.header_len)
	page.HeaderLen = int(cPage.header_len)
	page.Body = unsafe.Slice((*byte)(unsafe.Pointer(cPage.body)), cPage.body_len)
	page.BodyLen = int(cPage.body_len)

	return int(ret), nil
}

// NewOggStreamState 初始化Ogg流状态
func NewOggStreamState(serialno int) (*OggStreamState, error) {
	state := &OggStreamState{}
	ret := C.ogg_stream_init(&state.state, C.int(serialno))
	if ret != 0 {
		return nil, errors.New("failed to initialize ogg stream state")
	}
	return state, nil
}

// Clear 清理Ogg流状态
func (s *OggStreamState) Clear() error {
	if ret := C.ogg_stream_clear(&s.state); ret != 0 {
		return errors.New("failed to clear ogg stream state")
	}
	return nil
}

// PacketIn adds a packet to the stream
// The packet data must remain valid until the packet is no longer needed by libogg
func (s *OggStreamState) PacketIn(packet *OggPacket) error {
	var cPacket C.ogg_packet

	// Allocate memory for packet data
	cData := C.malloc(C.size_t(len(packet.Packet)))
	if cData == nil {
		return errors.New("failed to allocate memory for packet data")
	}
	defer C.free(cData) // Free memory when function returns

	// Copy data to C memory
	C.memcpy(cData, unsafe.Pointer(&packet.Packet[0]), C.size_t(len(packet.Packet)))

	// Set packet data
	cPacket.packet = (*C.uchar)(cData)
	cPacket.bytes = C.long(packet.Bytes)
	cPacket.b_o_s = C.long(packet.BOS)
	cPacket.e_o_s = C.long(packet.EOS)
	cPacket.granulepos = C.ogg_int64_t(packet.Granulepos)
	cPacket.packetno = C.ogg_int64_t(packet.Packetno)

	// Call C function
	ret := C.ogg_stream_packetin(&s.state, &cPacket)
	if ret != 0 {
		return errors.New("failed to add packet to stream")
	}

	// Note: The packet data is now owned by libogg and will be freed by it
	return nil
}

// PageOut extracts a page from the stream
func (s *OggStreamState) PageOut(page *OggPage) (int, error) {
	var cPage C.ogg_page
	ret := C.ogg_stream_pageout(&s.state, &cPage)
	if ret < 0 {
		return 0, errors.New("failed to get page from stream")
	}
	if ret == 0 {
		return 0, nil
	}

	// Check for nil pointers
	if cPage.header == nil || cPage.body == nil {
		return 0, errors.New("invalid page data")
	}

	// Convert C struct to Go struct
	page.Header = unsafe.Slice((*byte)(unsafe.Pointer(cPage.header)), cPage.header_len)
	page.HeaderLen = int(cPage.header_len)
	page.Body = unsafe.Slice((*byte)(unsafe.Pointer(cPage.body)), cPage.body_len)
	page.BodyLen = int(cPage.body_len)

	return int(ret), nil
}

// Flush forces pages to be written to the stream
func (s *OggStreamState) Flush(page *OggPage) (int, error) {
	var cPage C.ogg_page
	ret := C.ogg_stream_flush(&s.state, &cPage)
	if ret < 0 {
		return 0, errors.New("failed to flush stream")
	}
	if ret == 0 {
		return 0, nil
	}

	// Check for nil pointers
	if cPage.header == nil || cPage.body == nil {
		return 0, errors.New("invalid page data")
	}

	// Convert C struct to Go struct
	page.Header = unsafe.Slice((*byte)(unsafe.Pointer(cPage.header)), cPage.header_len)
	page.HeaderLen = int(cPage.header_len)
	page.Body = unsafe.Slice((*byte)(unsafe.Pointer(cPage.body)), cPage.body_len)
	page.BodyLen = int(cPage.body_len)

	return int(ret), nil
}

// PageIn 将页面添加到流中
func (s *OggStreamState) PageIn(page *OggPage) error {
	var cPage C.ogg_page
	cPage.header = (*C.uchar)(unsafe.Pointer(&page.Header[0]))
	cPage.header_len = C.long(page.HeaderLen)
	cPage.body = (*C.uchar)(unsafe.Pointer(&page.Body[0]))
	cPage.body_len = C.long(page.BodyLen)
	ret := C.ogg_stream_pagein(&s.state, &cPage)
	if ret != 0 {
		return errors.New("failed to add page to stream")
	}
	return nil
}

// PacketOut 从流中提取数据包
func (s *OggStreamState) PacketOut(packet *OggPacket) (int, error) {
	var cPacket C.ogg_packet
	ret := C.ogg_stream_packetout(&s.state, &cPacket)
	if ret < 0 {
		return 0, errors.New("failed to get packet from stream")
	}
	if ret == 0 {
		return 0, nil
	}
	packet.Packet = unsafe.Slice((*byte)(unsafe.Pointer(cPacket.packet)), cPacket.bytes)
	packet.Bytes = int(cPacket.bytes)
	packet.BOS = int(cPacket.b_o_s)
	packet.EOS = int(cPacket.e_o_s)
	packet.Granulepos = int64(cPacket.granulepos)
	packet.Packetno = int64(cPacket.packetno)
	return int(ret), nil
}
