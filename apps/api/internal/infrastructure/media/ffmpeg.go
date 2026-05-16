package media

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// SplitMP4ToFrames runs FFmpeg to extract every frame from an MP4 file as PNG images.
// mp4Path is the path to the input MP4 file. outputDir is the directory where PNG
// frames will be written as frame_0001.png, frame_0002.png, etc.
// Returns the list of output file paths in order.
func SplitMP4ToFrames(ctx context.Context, mp4Path string, outputDir string) ([]string, error) {
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return nil, fmt.Errorf("create output directory: %w", err)
	}

	outputPattern := filepath.Join(outputDir, "frame_%04d.png")
	args := []string{
		"-i", mp4Path,
		"-vframes", "0",
		"-f", "null", "-",
	}
	if err := runFFmpeg(ctx, nil, args...); err != nil {
		return nil, fmt.Errorf("probe input: %w", err)
	}

	args = []string{
		"-i", mp4Path,
		"-y",
		outputPattern,
	}
	if err := runFFmpeg(ctx, nil, args...); err != nil {
		return nil, fmt.Errorf("split to frames: %w", err)
	}

	entries, err := os.ReadDir(outputDir)
	if err != nil {
		return nil, fmt.Errorf("read output directory: %w", err)
	}

	var framePaths []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if strings.HasPrefix(entry.Name(), "frame_") && strings.HasSuffix(entry.Name(), ".png") {
			framePaths = append(framePaths, filepath.Join(outputDir, entry.Name()))
		}
	}

	if len(framePaths) == 0 {
		return nil, fmt.Errorf("no frames extracted from %s", mp4Path)
	}

	return framePaths, nil
}

// EncodeFramesToMP4 runs FFmpeg to encode a sequence of PNG frames into an MP4 video.
// framePaths must be sorted in display order. fps is the output frame rate.
// outputPath is the destination MP4 file path.
func EncodeFramesToMP4(ctx context.Context, framePaths []string, fps int, outputPath string) error {
	if len(framePaths) == 0 {
		return fmt.Errorf("no frames to encode")
	}
	if fps <= 0 {
		return fmt.Errorf("fps must be positive, got %d", fps)
	}

	tmpDir, err := os.MkdirTemp("", "anifusion-encode-*")
	if err != nil {
		return fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	inputListPath := filepath.Join(tmpDir, "input.txt")
	var listContent strings.Builder
	for i, path := range framePaths {
		linkPath := filepath.Join(tmpDir, fmt.Sprintf("frame_%04d.png", i+1))
		if err := os.Symlink(path, linkPath); err != nil {
			if err := copyFile(path, linkPath); err != nil {
				return fmt.Errorf("prepare frame %d: %w", i+1, err)
			}
		}
		listContent.WriteString(fmt.Sprintf("file '%s'\n", linkPath))
		listContent.WriteString(fmt.Sprintf("duration %.6f\n", 1.0/float64(fps)))
	}
	if err := os.WriteFile(inputListPath, []byte(listContent.String()), 0o644); err != nil {
		return fmt.Errorf("write input list: %w", err)
	}

	args := []string{
		"-f", "concat",
		"-safe", "0",
		"-i", inputListPath,
		"-c:v", "libx264",
		"-pix_fmt", "yuv420p",
		"-r", strconv.Itoa(fps),
		"-y",
		outputPath,
	}
	if err := runFFmpeg(ctx, nil, args...); err != nil {
		return fmt.Errorf("encode to MP4: %w", err)
	}

	return nil
}

func runFFmpeg(ctx context.Context, stdin []byte, args ...string) error {
	cmd := exec.CommandContext(ctx, "ffmpeg", args...)
	cmd.Stderr = &ffmpegStderr{}

	if len(stdin) > 0 {
		cmd.Stdin = strings.NewReader(string(stdin))
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ffmpeg: %w", err)
	}

	return nil
}

type ffmpegStderr struct {
	lines []string
}

func (w *ffmpegStderr) Write(p []byte) (int, error) {
	w.lines = append(w.lines, strings.TrimSpace(string(p)))
	return len(p), nil
}

func (w *ffmpegStderr) LastLine() string {
	if len(w.lines) == 0 {
		return ""
	}
	return w.lines[len(w.lines)-1]
}

func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0o644)
}
