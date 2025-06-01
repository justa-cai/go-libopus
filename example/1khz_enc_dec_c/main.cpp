#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <math.h>
#include <ogg/ogg.h>
#include <opus/opus.h>

#define SAMPLE_RATE 48000
#define CHANNELS 1
#define FRAME_SIZE 480  // 10ms at 48kHz
#define MAX_FRAME_SIZE 6*FRAME_SIZE
#define MAX_PACKET_SIZE (3*1276)
#define PI 3.14159265358979323846
#define DURATION_SECONDS 10
#define BITRATE 64000

// Function to generate 1kHz sine wave
void generate_sine_wave(opus_int16* buffer, int num_samples) {
    for (int i = 0; i < num_samples; i++) {
        buffer[i] = (opus_int16)(32767.0 * sin(2.0 * PI * 1000.0 * i / SAMPLE_RATE));
    }
}

// Function to write WAV header
void write_wav_header(FILE* file, int num_samples) {
    unsigned char header[44];
    int data_size = num_samples * sizeof(opus_int16);
    
    // RIFF header
    memcpy(header, "RIFF", 4);
    *(int*)(header + 4) = 36 + data_size;
    memcpy(header + 8, "WAVE", 4);
    
    // fmt chunk
    memcpy(header + 12, "fmt ", 4);
    *(int*)(header + 16) = 16;
    *(short*)(header + 20) = 1;  // PCM format
    *(short*)(header + 22) = CHANNELS;
    *(int*)(header + 24) = SAMPLE_RATE;
    *(int*)(header + 28) = SAMPLE_RATE * CHANNELS * sizeof(opus_int16);
    *(short*)(header + 32) = CHANNELS * sizeof(opus_int16);
    *(short*)(header + 34) = 16;  // bits per sample
    
    // data chunk
    memcpy(header + 36, "data", 4);
    *(int*)(header + 40) = data_size;
    
    fwrite(header, 1, 44, file);
}

// Function to write Ogg Opus header
void write_opus_header(ogg_stream_state* os, int serialno) {
    unsigned char header[19];
    ogg_packet op;

    // Magic signature
    memcpy(header, "OpusHead", 8);
    // Version
    header[8] = 1;
    // Channel count
    header[9] = CHANNELS;
    // Pre-skip
    header[10] = 0;
    header[11] = 0;
    // Sample rate
    header[12] = (SAMPLE_RATE >> 0) & 0xFF;
    header[13] = (SAMPLE_RATE >> 8) & 0xFF;
    header[14] = (SAMPLE_RATE >> 16) & 0xFF;
    header[15] = (SAMPLE_RATE >> 24) & 0xFF;
    // Output gain
    header[16] = 0;
    header[17] = 0;
    // Channel mapping
    header[18] = 0;

    op.packet = header;
    op.bytes = 19;
    op.b_o_s = 1;
    op.e_o_s = 0;
    op.granulepos = 0;
    op.packetno = 0;

    ogg_stream_packetin(os, &op);
}

// Function to write Ogg Opus comment header
void write_opus_comments(ogg_stream_state* os) {
    const char* vendor = "libopus 1.3.1";
    int vendor_length = strlen(vendor);
    unsigned char* header = (unsigned char*)malloc(8 + 4 + vendor_length + 4);
    ogg_packet op;

    // Magic signature
    memcpy(header, "OpusTags", 8);
    // Vendor string length
    header[8] = (vendor_length >> 0) & 0xFF;
    header[9] = (vendor_length >> 8) & 0xFF;
    header[10] = (vendor_length >> 16) & 0xFF;
    header[11] = (vendor_length >> 24) & 0xFF;
    // Vendor string
    memcpy(header + 12, vendor, vendor_length);
    // User comment list length (0)
    header[12 + vendor_length] = 0;
    header[13 + vendor_length] = 0;
    header[14 + vendor_length] = 0;
    header[15 + vendor_length] = 0;

    op.packet = header;
    op.bytes = 16 + vendor_length;
    op.b_o_s = 0;
    op.e_o_s = 0;
    op.granulepos = 0;
    op.packetno = 1;

    ogg_stream_packetin(os, &op);
    free(header);
}

// Function to encode sine wave to ogg file
int encode_sine_wave(const char* output_file) {
    printf("Starting encoding process...\n");
    printf("Generating %d seconds of 1kHz sine wave...\n", DURATION_SECONDS);

    // Create encoder
    int error;
    OpusEncoder* encoder = opus_encoder_create(SAMPLE_RATE, CHANNELS, OPUS_APPLICATION_AUDIO, &error);
    if (error != OPUS_OK) {
        fprintf(stderr, "Error: Failed to create encoder: %s\n", opus_strerror(error));
        return 1;
    }

    // Set encoder parameters
    error = opus_encoder_ctl(encoder, OPUS_SET_BITRATE(BITRATE));
    if (error != OPUS_OK) {
        fprintf(stderr, "Error: Failed to set bitrate: %s\n", opus_strerror(error));
        opus_encoder_destroy(encoder);
        return 1;
    }

    error = opus_encoder_ctl(encoder, OPUS_SET_COMPLEXITY(10));
    if (error != OPUS_OK) {
        fprintf(stderr, "Error: Failed to set complexity: %s\n", opus_strerror(error));
        opus_encoder_destroy(encoder);
        return 1;
    }

    // Create Ogg stream
    ogg_stream_state os;
    int serialno = rand();
    ogg_stream_init(&os, serialno);

    // Open output file
    FILE* outfile = fopen(output_file, "wb");
    if (!outfile) {
        fprintf(stderr, "Error: Failed to open output file: %s\n", output_file);
        opus_encoder_destroy(encoder);
        return 1;
    }

    // Write Ogg Opus headers
    write_opus_header(&os, serialno);
    write_opus_comments(&os);

    // Write Ogg pages for headers
    ogg_page og;
    while (ogg_stream_flush(&os, &og)) {
        fwrite(og.header, 1, og.header_len, outfile);
        fwrite(og.body, 1, og.body_len, outfile);
    }

    // Generate audio
    int total_samples = SAMPLE_RATE * DURATION_SECONDS;
    int num_frames = total_samples / FRAME_SIZE;
    
    // Allocate buffers
    opus_int16* pcm = (opus_int16*)malloc(FRAME_SIZE * sizeof(opus_int16));
    unsigned char* packet = (unsigned char*)malloc(MAX_PACKET_SIZE);
    if (!pcm || !packet) {
        fprintf(stderr, "Error: Memory allocation failed\n");
        goto cleanup;
    }

    ogg_packet op;

    printf("Encoding frames...\n");
    for (int i = 0; i < num_frames; i++) {
        // Show progress every 10%
        if (i % (num_frames / 10) == 0) {
            printf("Progress: %d%%\n", (i * 100) / num_frames);
        }

        // Generate sine wave for this frame
        generate_sine_wave(pcm, FRAME_SIZE);

        // Encode frame
        int nbBytes = opus_encode(encoder, pcm, FRAME_SIZE, packet, MAX_PACKET_SIZE);
        if (nbBytes < 0) {
            fprintf(stderr, "Error: Failed to encode frame: %s\n", opus_strerror(nbBytes));
            goto cleanup;
        }

        // Create Ogg packet
        op.packet = packet;
        op.bytes = nbBytes;
        op.b_o_s = 0;
        op.e_o_s = (i == num_frames - 1);
        op.granulepos = (i + 1) * FRAME_SIZE;
        op.packetno = i + 2;  // +2 because we already used 0 and 1 for headers

        // Add packet to Ogg stream
        if (ogg_stream_packetin(&os, &op) != 0) {
            fprintf(stderr, "Error: Failed to add packet to Ogg stream\n");
            goto cleanup;
        }

        // Write Ogg pages
        while (ogg_stream_pageout(&os, &og)) {
            fwrite(og.header, 1, og.header_len, outfile);
            fwrite(og.body, 1, og.body_len, outfile);
        }
    }

    // Flush remaining pages
    while (ogg_stream_flush(&os, &og)) {
        fwrite(og.header, 1, og.header_len, outfile);
        fwrite(og.body, 1, og.body_len, outfile);
    }

    printf("Encoding completed successfully!\n");
    printf("Output saved to: %s\n", output_file);

cleanup:
    // Cleanup
    free(pcm);
    free(packet);
    fclose(outfile);
    ogg_stream_clear(&os);
    opus_encoder_destroy(encoder);
    return 0;
}

// Function to decode ogg file to wav
int decode_ogg_to_wav(const char* input_file, const char* output_file) {
    printf("Starting decoding process...\n");
    printf("Input file: %s\n", input_file);

    // Create decoder
    int error;
    OpusDecoder* decoder = opus_decoder_create(SAMPLE_RATE, CHANNELS, &error);
    if (error != OPUS_OK) {
        fprintf(stderr, "Error: Failed to create decoder: %s\n", opus_strerror(error));
        return 1;
    }

    // Open input file
    FILE* infile = fopen(input_file, "rb");
    if (!infile) {
        fprintf(stderr, "Error: Failed to open input file: %s\n", input_file);
        opus_decoder_destroy(decoder);
        return 1;
    }

    // Initialize Ogg sync state
    ogg_sync_state oy;
    ogg_sync_init(&oy);

    // Initialize Ogg stream state
    ogg_stream_state os;
    ogg_page og;
    ogg_packet op;

    // Open output file
    FILE* outfile = fopen(output_file, "wb");
    if (!outfile) {
        fprintf(stderr, "Error: Failed to open output file: %s\n", output_file);
        fclose(infile);
        opus_decoder_destroy(decoder);
        return 1;
    }

    // Buffer for decoded audio
    opus_int16* pcm = (opus_int16*)malloc(MAX_FRAME_SIZE * sizeof(opus_int16));
    if (!pcm) {
        fprintf(stderr, "Error: Memory allocation failed\n");
        goto cleanup;
    }

    int total_samples = 0;
    int serialno = -1;
    int frame_count = 0;

    printf("Decoding frames...\n");
    // Read and decode
    while (1) {
        // Read a page
        char* buffer = ogg_sync_buffer(&oy, 4096);
        int bytes = fread(buffer, 1, 4096, infile);
        if (bytes == 0) break;
        ogg_sync_wrote(&oy, bytes);

        while (ogg_sync_pageout(&oy, &og) == 1) {
            if (serialno == -1) {
                serialno = ogg_page_serialno(&og);
                ogg_stream_init(&os, serialno);
            }

            if (ogg_stream_pagein(&os, &og) != 0) {
                fprintf(stderr, "Error: Failed to read Ogg page\n");
                goto cleanup;
            }

            while (ogg_stream_packetout(&os, &op) == 1) {
                // Decode packet
                int frame_size = opus_decode(decoder, op.packet, op.bytes, pcm, MAX_FRAME_SIZE, 0);
                if (frame_size < 0) {
                    fprintf(stderr, "Error: Failed to decode packet: %s\n", opus_strerror(frame_size));
                    goto cleanup;
                }

                // Write decoded audio
                fwrite(pcm, sizeof(opus_int16), frame_size, outfile);
                total_samples += frame_size;
                frame_count++;

                // Show progress every 100 frames
                if (frame_count % 100 == 0) {
                    printf("Decoded %d frames...\r", frame_count);
                    fflush(stdout);
                }
            }
        }
    }

    // Write WAV header
    rewind(outfile);
    write_wav_header(outfile, total_samples);

    printf("\nDecoding completed successfully!\n");
    printf("Output saved to: %s\n", output_file);
    printf("Total samples decoded: %d\n", total_samples);

cleanup:
    // Cleanup
    free(pcm);
    fclose(infile);
    fclose(outfile);
    ogg_stream_clear(&os);
    ogg_sync_clear(&oy);
    opus_decoder_destroy(decoder);
    return 0;
}

int main() {
    const char* raw_file = "sine_raw.wav";
    const char* encoded_file = "sine_encoded.ogg";
    const char* decoded_file = "sine_decoded.wav";

    printf("Step 1: Generating 1kHz sine wave...\n");
    // Generate raw sine wave and save as WAV
    int total_samples = SAMPLE_RATE * DURATION_SECONDS;
    opus_int16* raw_pcm = (opus_int16*)malloc(total_samples * sizeof(opus_int16));
    if (!raw_pcm) {
        fprintf(stderr, "Error: Memory allocation failed\n");
        return 1;
    }

    // Generate sine wave
    generate_sine_wave(raw_pcm, total_samples);

    // Save as WAV
    FILE* raw_out = fopen(raw_file, "wb");
    if (!raw_out) {
        fprintf(stderr, "Error: Failed to open output file: %s\n", raw_file);
        free(raw_pcm);
        return 1;
    }
    write_wav_header(raw_out, total_samples);
    fwrite(raw_pcm, sizeof(opus_int16), total_samples, raw_out);
    fclose(raw_out);
    printf("Raw sine wave saved to: %s\n", raw_file);

    printf("\nStep 2: Encoding to Ogg/Opus...\n");
    if (encode_sine_wave(encoded_file) != 0) {
        fprintf(stderr, "Error: Encoding failed\n");
        free(raw_pcm);
        return 1;
    }

    printf("\nStep 3: Decoding back to WAV...\n");
    if (decode_ogg_to_wav(encoded_file, decoded_file) != 0) {
        fprintf(stderr, "Error: Decoding failed\n");
        free(raw_pcm);
        return 1;
    }

    printf("\nAll steps completed successfully!\n");
    printf("Files generated:\n");
    printf("1. Raw sine wave: %s\n", raw_file);
    printf("2. Encoded file: %s\n", encoded_file);
    printf("3. Decoded file: %s\n", decoded_file);

    free(raw_pcm);
    return 0;
}
