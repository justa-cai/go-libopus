package main

import (
	"encoding/binary"
	"fmt"
	"math"
	"os"

	"github.com/justa-cai/go-libopus/ogg"
	"github.com/justa-cai/go-libopus/opus"
)

const (
	sampleRate    = 48000
	channels      = 1
	duration      = 10   // seconds
	frequency     = 1000 // 1kHz
	frameSize     = 480  // 10ms at 48kHz
	maxFrameSize  = 6 * frameSize
	maxPacketSize = 3 * 1276
	bitrate       = 64000
)

// generateSineWave generates a 1kHz sine wave for the specified duration
func generateSineWave() []int16 {
	samples := sampleRate * duration
	data := make([]int16, samples)

	for i := 0; i < samples; i++ {
		// Generate sine wave
		data[i] = int16(math.Sin(2*math.Pi*float64(i)*frequency/float64(sampleRate)) * 32767)
	}

	return data
}

// writeWavHeader writes a WAV file header
func writeWavHeader(file *os.File, numSamples int) error {
	// RIFF header
	if _, err := file.Write([]byte("RIFF")); err != nil {
		return err
	}
	dataSize := numSamples * 2 // 2 bytes per sample
	if err := binary.Write(file, binary.LittleEndian, uint32(36+dataSize)); err != nil {
		return err
	}
	if _, err := file.Write([]byte("WAVE")); err != nil {
		return err
	}

	// fmt chunk
	if _, err := file.Write([]byte("fmt ")); err != nil {
		return err
	}
	if err := binary.Write(file, binary.LittleEndian, uint32(16)); err != nil {
		return err
	}
	if err := binary.Write(file, binary.LittleEndian, uint16(1)); err != nil { // PCM format
		return err
	}
	if err := binary.Write(file, binary.LittleEndian, uint16(channels)); err != nil {
		return err
	}
	if err := binary.Write(file, binary.LittleEndian, uint32(sampleRate)); err != nil {
		return err
	}
	if err := binary.Write(file, binary.LittleEndian, uint32(sampleRate*channels*2)); err != nil {
		return err
	}
	if err := binary.Write(file, binary.LittleEndian, uint16(channels*2)); err != nil {
		return err
	}
	if err := binary.Write(file, binary.LittleEndian, uint16(16)); err != nil {
		return err
	}

	// data chunk
	if _, err := file.Write([]byte("data")); err != nil {
		return err
	}
	if err := binary.Write(file, binary.LittleEndian, uint32(dataSize)); err != nil {
		return err
	}

	return nil
}

// writeOpusHeader writes the Opus header packet
func writeOpusHeader(stream *ogg.OggStreamState) error {
	// OpusHead header format:
	// - Magic signature: "OpusHead" (8 bytes)
	// - Version number: 1 (1 byte)
	// - Channel count: 1 (1 byte)
	// - Pre-skip: 0 (2 bytes, little endian)
	// - Input sample rate: 48000 (4 bytes, little endian)
	// - Output gain: 0 (2 bytes, little endian)
	// - Channel mapping family: 0 (1 byte)
	header := make([]byte, 19)

	// Magic signature
	copy(header[0:8], []byte("OpusHead"))

	// Version number
	header[8] = 1

	// Channel count
	header[9] = byte(channels)

	// Pre-skip (0)
	binary.LittleEndian.PutUint16(header[10:12], 0)

	// Input sample rate (48000)
	binary.LittleEndian.PutUint32(header[12:16], uint32(sampleRate))

	// Output gain (0)
	binary.LittleEndian.PutUint16(header[16:18], 0)

	// Channel mapping family (0)
	header[18] = 0

	packet := &ogg.OggPacket{
		Packet:     header,
		Bytes:      len(header),
		BOS:        1,
		EOS:        0,
		Granulepos: 0,
		Packetno:   0,
	}

	return stream.PacketIn(packet)
}

// writeOpusComments writes the Opus comments packet
func writeOpusComments(stream *ogg.OggStreamState) error {
	// OpusTags header format:
	// - Magic signature: "OpusTags" (8 bytes)
	// - Vendor string length (4 bytes, little endian)
	// - Vendor string (variable length)
	// - User comment list length (4 bytes, little endian)
	// - User comments (variable length)
	vendor := "go-libopus 1.0.0"
	vendorLength := len(vendor)

	// Calculate total size
	totalSize := 8 + 4 + vendorLength + 4
	header := make([]byte, totalSize)

	// Magic signature
	copy(header[0:8], []byte("OpusTags"))

	// Vendor string length
	binary.LittleEndian.PutUint32(header[8:12], uint32(vendorLength))

	// Vendor string
	copy(header[12:12+vendorLength], []byte(vendor))

	// User comment list length (0)
	binary.LittleEndian.PutUint32(header[12+vendorLength:], 0)

	packet := &ogg.OggPacket{
		Packet:     header,
		Bytes:      len(header),
		BOS:        0,
		EOS:        0,
		Granulepos: 0,
		Packetno:   1,
	}

	return stream.PacketIn(packet)
}

// encodeAndSave encodes the audio data and saves it as OGG
func encodeAndSave(audioData []int16, filename string) error {
	// Create Opus encoder
	encoder, err := opus.NewEncoder(sampleRate, channels, opus.OpusApplicationAudio)
	if err != nil {
		return fmt.Errorf("failed to create encoder: %v", err)
	}
	defer encoder.Close()

	// Set encoder parameters
	if err := encoder.SetBitrate(bitrate); err != nil {
		return fmt.Errorf("failed to set bitrate: %v", err)
	}
	if err := encoder.SetComplexity(10); err != nil {
		return fmt.Errorf("failed to set complexity: %v", err)
	}

	// Create OGG stream
	stream, err := ogg.NewOggStreamState(1)
	if err != nil {
		return fmt.Errorf("failed to create ogg stream: %v", err)
	}
	defer stream.Clear()

	// Create output file
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create output file: %v", err)
	}
	defer file.Close()

	// Write headers
	if err := writeOpusHeader(stream); err != nil {
		return fmt.Errorf("failed to write opus header: %v", err)
	}

	// Write header pages
	for {
		page := &ogg.OggPage{}
		ret, err := stream.PageOut(page)
		if err != nil {
			return fmt.Errorf("failed to get page: %v", err)
		}
		if ret == 0 {
			break
		}
		if _, err := file.Write(page.Header); err != nil {
			return fmt.Errorf("failed to write page header: %v", err)
		}
		if _, err := file.Write(page.Body); err != nil {
			return fmt.Errorf("failed to write page body: %v", err)
		}
	}

	// Write comments
	if err := writeOpusComments(stream); err != nil {
		return fmt.Errorf("failed to write opus comments: %v", err)
	}

	// Write comment pages
	for {
		page := &ogg.OggPage{}
		ret, err := stream.PageOut(page)
		if err != nil {
			return fmt.Errorf("failed to get page: %v", err)
		}
		if ret == 0 {
			break
		}
		if _, err := file.Write(page.Header); err != nil {
			return fmt.Errorf("failed to write page header: %v", err)
		}
		if _, err := file.Write(page.Body); err != nil {
			return fmt.Errorf("failed to write page body: %v", err)
		}
	}

	// Encode frames
	numFrames := len(audioData) / frameSize
	packetNo := int64(2)
	totalSamples := 0

	for i := 0; i < numFrames; i++ {
		start := i * frameSize
		end := start + frameSize
		if end > len(audioData) {
			end = len(audioData)
		}

		// Convert int16 samples to bytes
		frameData := make([]byte, (end-start)*2)
		for j := 0; j < end-start; j++ {
			binary.LittleEndian.PutUint16(frameData[j*2:], uint16(audioData[start+j]))
		}

		// Encode frame
		encodedData := make([]byte, maxPacketSize)
		encodedSize, err := encoder.Encode(frameData, encodedData)
		if err != nil {
			return fmt.Errorf("failed to encode frame: %v", err)
		}

		// Create OGG packet
		packet := &ogg.OggPacket{
			Packet:     encodedData[:encodedSize],
			Bytes:      encodedSize,
			BOS:        0,
			EOS:        int(boolToInt(i == numFrames-1)),
			Granulepos: int64(totalSamples),
			Packetno:   packetNo,
		}
		packetNo++
		totalSamples += frameSize

		// Add packet to stream
		if err := stream.PacketIn(packet); err != nil {
			return fmt.Errorf("failed to add packet to stream: %v", err)
		}

		// Write pages
		for {
			page := &ogg.OggPage{}
			ret, err := stream.PageOut(page)
			if err != nil {
				return fmt.Errorf("failed to get page: %v", err)
			}
			if ret == 0 {
				break
			}
			if _, err := file.Write(page.Header); err != nil {
				return fmt.Errorf("failed to write page header: %v", err)
			}
			if _, err := file.Write(page.Body); err != nil {
				return fmt.Errorf("failed to write page body: %v", err)
			}
		}
	}

	// Flush remaining pages
	for {
		page := &ogg.OggPage{}
		ret, err := stream.Flush(page)
		if err != nil {
			return fmt.Errorf("failed to flush stream: %v", err)
		}
		if ret == 0 {
			break
		}
		if _, err := file.Write(page.Header); err != nil {
			return fmt.Errorf("failed to write page header: %v", err)
		}
		if _, err := file.Write(page.Body); err != nil {
			return fmt.Errorf("failed to write page body: %v", err)
		}
	}

	return nil
}

// decodeAndSave decodes an OGG file and saves it as WAV
func decodeAndSave(inputFile, outputFile string) error {
	// Open input file
	file, err := os.Open(inputFile)
	if err != nil {
		return fmt.Errorf("failed to open input file: %v", err)
	}
	defer file.Close()

	// Create Opus decoder
	decoder, err := opus.NewDecoder(sampleRate, channels)
	if err != nil {
		return fmt.Errorf("failed to create decoder: %v", err)
	}
	defer decoder.Close()

	// Create OGG sync state
	sync, err := ogg.NewOggSyncState()
	if err != nil {
		return fmt.Errorf("failed to create sync state: %v", err)
	}
	defer sync.Clear()

	// Create output WAV file
	outFile, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("failed to create output file: %v", err)
	}
	defer outFile.Close()

	// Write WAV header (will be updated later)
	if err := writeWavHeader(outFile, 0); err != nil {
		return fmt.Errorf("failed to write WAV header: %v", err)
	}

	// Read and process OGG pages
	buffer := make([]byte, 4096)
	packetCount := 0
	totalSamples := 0
	var stream *ogg.OggStreamState

	for {
		// Read data into buffer
		n, err := file.Read(buffer)
		if err != nil {
			break
		}

		// Get buffer from sync state
		oggBuffer, err := sync.Buffer(n)
		if err != nil {
			return fmt.Errorf("failed to get buffer: %v", err)
		}

		// Copy data to OGG buffer
		copy(oggBuffer, buffer[:n])

		// Mark bytes as written
		if err := sync.Wrote(n); err != nil {
			return fmt.Errorf("failed to mark bytes as written: %v", err)
		}

		// Process pages
		for {
			page := &ogg.OggPage{}
			ret, err := sync.PageOut(page)
			if err != nil {
				return fmt.Errorf("failed to get page: %v", err)
			}
			if ret == 0 {
				break
			}

			// Initialize stream if needed
			if stream == nil {
				stream, err = ogg.NewOggStreamState(1)
				if err != nil {
					return fmt.Errorf("failed to create stream: %v", err)
				}
				defer stream.Clear()
			}

			// Add page to stream
			if err := stream.PageIn(page); err != nil {
				return fmt.Errorf("failed to add page to stream: %v", err)
			}

			// Process packets
			for {
				packet := &ogg.OggPacket{}
				ret, err := stream.PacketOut(packet)
				if err != nil {
					return fmt.Errorf("failed to get packet: %v", err)
				}
				if ret == 0 {
					break
				}

				// Skip header packets
				if packetCount < 2 {
					packetCount++
					continue
				}

				// Decode packet
				decodedData := make([]byte, maxFrameSize*2)
				decodedSize, err := decoder.Decode(packet.Packet, decodedData)
				if err != nil {
					return fmt.Errorf("failed to decode data: %v", err)
				}

				// Write decoded data
				if _, err := outFile.Write(decodedData[:decodedSize*2]); err != nil {
					return fmt.Errorf("failed to write decoded data: %v", err)
				}
				totalSamples += decodedSize
			}
		}
	}

	// Update WAV header with correct size
	if _, err := outFile.Seek(0, 0); err != nil {
		return fmt.Errorf("failed to seek to start of file: %v", err)
	}
	if err := writeWavHeader(outFile, totalSamples); err != nil {
		return fmt.Errorf("failed to update WAV header: %v", err)
	}

	return nil
}

// boolToInt converts a boolean to an integer (0 or 1)
func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func main() {
	// Generate sine wave
	fmt.Println("Generating 1kHz sine wave...")
	audioData := generateSineWave()

	// Save original WAV file
	fmt.Println("Saving original WAV file...")
	rawFile, err := os.Create("sine_raw.wav")
	if err != nil {
		fmt.Printf("Error creating raw WAV file: %v\n", err)
		return
	}
	defer rawFile.Close()

	if err := writeWavHeader(rawFile, len(audioData)); err != nil {
		fmt.Printf("Error writing WAV header: %v\n", err)
		return
	}

	// Convert int16 samples to bytes
	rawData := make([]byte, len(audioData)*2)
	for i, sample := range audioData {
		binary.LittleEndian.PutUint16(rawData[i*2:], uint16(sample))
	}
	if _, err := rawFile.Write(rawData); err != nil {
		fmt.Printf("Error writing raw data: %v\n", err)
		return
	}

	// Encode and save as OGG
	fmt.Println("Encoding and saving as OGG...")
	if err := encodeAndSave(audioData, "sine_encoded.ogg"); err != nil {
		fmt.Printf("Error encoding: %v\n", err)
		return
	}

	// Decode OGG file
	fmt.Println("Decoding OGG file...")
	if err := decodeAndSave("sine_encoded.ogg", "sine_decoded.wav"); err != nil {
		fmt.Printf("Error decoding: %v\n", err)
		return
	}

	fmt.Println("Done! Generated files:")
	fmt.Println("- sine_raw.wav: Original sine wave")
	fmt.Println("- sine_encoded.ogg: Opus encoded file")
	fmt.Println("- sine_decoded.wav: Decoded WAV file")
}
