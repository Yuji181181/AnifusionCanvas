package media

import (
	"context"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

func requireFFmpeg(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		t.Skip("ffmpeg not available")
	}
}

func TestSplitMP4ToFrames(t *testing.T) {
	requireFFmpeg(t)
	tmpDir := t.TempDir()
	mp4Path := filepath.Join(tmpDir, "test.mp4")

	if err := generateTestMP4(t, mp4Path, 1, 640, 360); err != nil {
		t.Skipf("cannot generate test MP4 (ffmpeg not available?): %v", err)
	}

	outputDir := filepath.Join(tmpDir, "frames")
	frames, err := SplitMP4ToFrames(context.Background(), mp4Path, outputDir)
	if err != nil {
		t.Fatalf("split MP4 to frames failed: %v", err)
	}
	if len(frames) != 1 {
		t.Fatalf("expected 1 frame, got %d", len(frames))
	}

	data, err := os.ReadFile(frames[0])
	if err != nil {
		t.Fatalf("read frame file: %v", err)
	}
	if len(data) < 100 {
		t.Fatalf("expected frame to have image data, got %d bytes", len(data))
	}
}

func TestSplitMP4ToFramesMultipleFrames(t *testing.T) {
	requireFFmpeg(t)
	tmpDir := t.TempDir()
	mp4Path := filepath.Join(tmpDir, "test.mp4")

	if err := generateTestMP4(t, mp4Path, 3, 640, 360); err != nil {
		t.Skipf("cannot generate test MP4: %v", err)
	}

	outputDir := filepath.Join(tmpDir, "frames")
	frames, err := SplitMP4ToFrames(context.Background(), mp4Path, outputDir)
	if err != nil {
		t.Fatalf("split MP4 to frames failed: %v", err)
	}
	if len(frames) < 2 {
		t.Fatalf("expected at least 2 frames, got %d", len(frames))
	}
}

func TestEncodeFramesToMP4(t *testing.T) {
	requireFFmpeg(t)
	tmpDir := t.TempDir()

	framePath := filepath.Join(tmpDir, "frame_0001.png")
	if err := generateTestPNG(t, framePath, 640, 360); err != nil {
		t.Fatalf("generate test PNG: %v", err)
	}

	outputPath := filepath.Join(tmpDir, "output.mp4")
	if err := EncodeFramesToMP4(context.Background(), []string{framePath}, 12, outputPath); err != nil {
		t.Fatalf("encode frames to MP4 failed: %v", err)
	}

	stat, err := os.Stat(outputPath)
	if err != nil {
		t.Fatalf("stat output file: %v", err)
	}
	if stat.Size() < 100 {
		t.Fatalf("expected output MP4 to have some data, got %d bytes", stat.Size())
	}
}

func TestEncodeFramesToMP4MultipleFiles(t *testing.T) {
	requireFFmpeg(t)
	tmpDir := t.TempDir()

	var framePaths []string
	for i := 1; i <= 3; i++ {
		path := filepath.Join(tmpDir, "frame_"+padZero(i, 4)+".png")
		if err := generateTestPNG(t, path, 640, 360); err != nil {
			t.Fatalf("generate test PNG %d: %v", i, err)
		}
		framePaths = append(framePaths, path)
	}

	outputPath := filepath.Join(tmpDir, "output.mp4")
	if err := EncodeFramesToMP4(context.Background(), framePaths, 12, outputPath); err != nil {
		t.Fatalf("encode frames to MP4 failed: %v", err)
	}
}

func TestSplitMP4ToFramesNonexistentInput(t *testing.T) {
	_, err := SplitMP4ToFrames(context.Background(), "/nonexistent/path.mp4", t.TempDir())
	if err == nil {
		t.Fatalf("expected error for nonexistent input")
	}
}

func TestEncodeFramesToMP4NoFrames(t *testing.T) {
	err := EncodeFramesToMP4(context.Background(), nil, 12, filepath.Join(t.TempDir(), "out.mp4"))
	if err == nil {
		t.Fatalf("expected error for empty frame list")
	}
}

func TestEncodeFramesToMP4InvalidFPS(t *testing.T) {
	err := EncodeFramesToMP4(context.Background(), []string{"frame.png"}, 0, filepath.Join(t.TempDir(), "out.mp4"))
	if err == nil {
		t.Fatalf("expected error for fps=0")
	}
}

func TestEncodeFramesToMP4ContextCancel(t *testing.T) {
	requireFFmpeg(t)
	tmpDir := t.TempDir()
	framePath := filepath.Join(tmpDir, "frame_0001.png")
	if err := generateTestPNG(t, framePath, 640, 360); err != nil {
		t.Fatalf("generate test PNG: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := EncodeFramesToMP4(ctx, []string{framePath}, 12, filepath.Join(tmpDir, "out.mp4"))
	if err == nil {
		t.Fatalf("expected error for canceled context")
	}
}

func TestSplitMP4ToFramesContextCancel(t *testing.T) {
	requireFFmpeg(t)
	tmpDir := t.TempDir()
	mp4Path := filepath.Join(tmpDir, "test.mp4")

	if err := generateTestMP4(t, mp4Path, 1, 640, 360); err != nil {
		t.Skipf("cannot generate test MP4: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()
	time.Sleep(1 * time.Millisecond)

	_, err := SplitMP4ToFrames(ctx, mp4Path, filepath.Join(tmpDir, "frames"))
	if err == nil {
		t.Fatalf("expected error for expired context")
	}
}

func generateTestMP4(t *testing.T, outputPath string, frameCount int, width, height int) error {
	t.Helper()

	tmpDir := t.TempDir()

	var framePaths []string
	for i := 1; i <= frameCount; i++ {
		path := filepath.Join(tmpDir, "frame_"+padZero(i, 4)+".png")
		if err := generateTestPNG(t, path, width, height); err != nil {
			return err
		}
		framePaths = append(framePaths, path)
	}

	inputListPath := filepath.Join(tmpDir, "input.txt")
	var content string
	for _, p := range framePaths {
		content += "file '" + p + "'\n"
		content += "duration 0.25\n"
	}
	if err := os.WriteFile(inputListPath, []byte(content), 0o644); err != nil {
		return err
	}

	cmd := exec.Command("ffmpeg",
		"-f", "concat", "-safe", "0", "-i", inputListPath,
		"-c:v", "libx264", "-pix_fmt", "yuv420p", "-r", "4",
		"-y", outputPath,
	)
	return cmd.Run()
}

func generateTestPNG(t *testing.T, outputPath string, width, height int) error {
	t.Helper()

	img := image.NewRGBA(image.Rect(0, 0, width, height))
	c := color.RGBA{R: 100, G: 150, B: 200, A: 255}
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.SetRGBA(x, y, c)
		}
	}

	f, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer f.Close()

	return png.Encode(f, img)
}

func padZero(n, width int) string {
	return fmt.Sprintf("%0*d", width, n)
}
