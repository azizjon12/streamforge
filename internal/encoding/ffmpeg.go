package encoding

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
)

type FFmpeg struct {
	Binary string // default "ffmpeg"
	Preset HLSPreset
}

func NewFFmpeg(binary string, preset HLSPreset) *FFmpeg {
	if binary == "" {
		binary = "ffmpeg"
	}
	return &FFmpeg{Binary: binary, Preset: preset}
}

// ToHLS converts input file -> HLS in outputDir. Returns playlist filename (absolute path)
func (f *FFmpeg) ToHLS(ctx context.Context, inputPath, outputDir string) (string, error) {
	playlist := filepath.Join(outputDir, "playlist.m3u8")
	segmentPattern := filepath.Join(outputDir, "segment_%03d.ts")

	args := []string{
		"-y",
		"-i", inputPath,
		"-codec:v", f.Preset.VideoCodec,
		"-codec:a", f.Preset.AudioCodec,
		"-preset", f.Preset.Preset,
		"-g", fmt.Sprintf("%d", f.Preset.GOP),
		"-sc_threshold", "0",
		"-hls_time", fmt.Sprintf("%d", f.Preset.SegmentSeconds),
		"-hls_playlist_type", "vod",
		"-hls_segment_filename", segmentPattern,
		playlist,
	}

	cmd := exec.CommandContext(ctx, f.Binary, args...)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("ffmpeg failed: %w; stderr: %s", err, stderr.String())
	}
	return playlist, nil
}
