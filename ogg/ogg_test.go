package ogg_test

import (
	"testing"

	"github.com/justa-cai/go-libopus/ogg"
)

func TestNewOggSyncState(t *testing.T) {
	state, err := ogg.NewOggSyncState()
	if err != nil {
		t.Fatalf("Failed to create OggSyncState: %v", err)
	}
	if state == nil {
		t.Fatal("OggSyncState is nil")
	}
	defer state.Clear()
}

func TestBufferOperations(t *testing.T) {
	state, err := ogg.NewOggSyncState()
	if err != nil {
		t.Fatalf("Failed to create OggSyncState: %v", err)
	}
	defer state.Clear()

	// Test buffer allocation
	bufferSize := 4096
	buffer, err := state.Buffer(bufferSize)
	if err != nil {
		t.Fatalf("Failed to get buffer: %v", err)
	}
	if len(buffer) != bufferSize {
		t.Errorf("Buffer size mismatch: got %d, want %d", len(buffer), bufferSize)
	}

	// Test writing to buffer
	err = state.Wrote(bufferSize)
	if err != nil {
		t.Fatalf("Failed to mark written bytes: %v", err)
	}
}

func TestPageOut(t *testing.T) {
	state, err := ogg.NewOggSyncState()
	if err != nil {
		t.Fatalf("Failed to create OggSyncState: %v", err)
	}
	defer state.Clear()

	// Create a page structure
	page := &ogg.OggPage{}

	// Try to get a page (should return 0 as no data has been written)
	ret, err := state.PageOut(page)
	if err != nil {
		t.Fatalf("Failed to get page: %v", err)
	}
	if ret != 0 {
		t.Errorf("Expected 0 pages, got %d", ret)
	}
}

func TestClear(t *testing.T) {
	state, err := ogg.NewOggSyncState()
	if err != nil {
		t.Fatalf("Failed to create OggSyncState: %v", err)
	}

	// Test clearing the state
	err = state.Clear()
	if err != nil {
		t.Fatalf("Failed to clear state: %v", err)
	}
}

func TestNewOggStreamState(t *testing.T) {
	serialno := 12345
	state, err := ogg.NewOggStreamState(serialno)
	if err != nil {
		t.Fatalf("Failed to create OggStreamState: %v", err)
	}
	if state == nil {
		t.Fatal("OggStreamState is nil")
	}
	defer state.Clear()
}

func TestPacketIn(t *testing.T) {
	state, err := ogg.NewOggStreamState(12345)
	if err != nil {
		t.Fatalf("Failed to create OggStreamState: %v", err)
	}
	defer state.Clear()

	// Create a test packet with fixed data
	data := make([]byte, 1024)
	for i := range data {
		data[i] = byte(i % 256)
	}

	// Create a test packet
	packet := &ogg.OggPacket{
		Packet:     data,
		Bytes:      len(data),
		BOS:        1,
		EOS:        0,
		Granulepos: 0,
		Packetno:   0,
	}

	// Test adding packet to stream
	err = state.PacketIn(packet)
	if err != nil {
		t.Fatalf("Failed to add packet to stream: %v", err)
	}
}

func TestStreamPageOut(t *testing.T) {
	state, err := ogg.NewOggStreamState(12345)
	if err != nil {
		t.Fatalf("Failed to create OggStreamState: %v", err)
	}
	defer state.Clear()

	// Create a page structure
	page := &ogg.OggPage{}

	// Try to get a page (should return 0 as no packets have been added)
	ret, err := state.PageOut(page)
	if err != nil {
		t.Fatalf("Failed to get page from stream: %v", err)
	}
	if ret != 0 {
		t.Errorf("Expected 0 pages, got %d", ret)
	}
}

func TestStreamFlush(t *testing.T) {
	state, err := ogg.NewOggStreamState(12345)
	if err != nil {
		t.Fatalf("Failed to create OggStreamState: %v", err)
	}
	defer state.Clear()

	// Create a page structure
	page := &ogg.OggPage{}

	// Try to flush the stream (should return 0 as no packets have been added)
	ret, err := state.Flush(page)
	if err != nil {
		t.Fatalf("Failed to flush stream: %v", err)
	}
	if ret != 0 {
		t.Errorf("Expected 0 pages, got %d", ret)
	}
}

func TestStreamClear(t *testing.T) {
	state, err := ogg.NewOggStreamState(12345)
	if err != nil {
		t.Fatalf("Failed to create OggStreamState: %v", err)
	}

	// Test clearing the state
	err = state.Clear()
	if err != nil {
		t.Fatalf("Failed to clear stream state: %v", err)
	}
}
