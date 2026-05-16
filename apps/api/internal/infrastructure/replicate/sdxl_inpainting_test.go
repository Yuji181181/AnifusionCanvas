package replicate

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/color"
	"image/png"
	"strings"
	"testing"

	"github.com/haseg/anifusion-canvas/apps/api/internal/domain"
)

func TestBuildSDXLInpaintingInput(t *testing.T) {
	maskDataURL := buildTestMaskDataURL(t, 4, 4, func(x, y int) color.Color {
		if x < 2 && y < 2 {
			return color.Black
		}
		return color.Transparent
	})

	req := domain.InpaintFrameRequest{
		ProjectID:   "proj-1",
		FrameID:     "frame-1",
		Prompt:      "fix the hand",
		MaskDataURL: maskDataURL,
		Strength:    0.8,
	}

	input, err := BuildSDXLInpaintingInput(req, "data:image/png;base64,frameImage")
	if err != nil {
		t.Fatalf("build input failed: %v", err)
	}

	if input["image"] != "data:image/png;base64,frameImage" {
		t.Fatalf("unexpected frame image: %v", input["image"])
	}
	if input["prompt"] != "fix the hand" {
		t.Fatalf("unexpected prompt: %v", input["prompt"])
	}
	if input["strength"] != 0.8 {
		t.Fatalf("unexpected strength: %v", input["strength"])
	}
	if input["num_inference_steps"] != defaultInpaintingSteps {
		t.Fatalf("unexpected steps: %v", input["num_inference_steps"])
	}

	maskStr, ok := input["mask"].(string)
	if !ok {
		t.Fatalf("expected mask to be a string data URL")
	}
	if !strings.HasPrefix(maskStr, "data:image/png;base64,") {
		t.Fatalf("expected mask to be a PNG data URL")
	}
}

func TestMaskInversion(t *testing.T) {
	maskDataURL := buildTestMaskDataURL(t, 4, 4, func(x, y int) color.Color {
		if x < 2 && y < 2 {
			return color.Black
		}
		return color.Transparent
	})

	req := domain.InpaintFrameRequest{
		ProjectID:   "proj-1",
		FrameID:     "frame-1",
		Prompt:      "fix",
		MaskDataURL: maskDataURL,
		Strength:    0.5,
	}

	input, err := BuildSDXLInpaintingInput(req, "data:image/png;base64,frameImage")
	if err != nil {
		t.Fatalf("build input failed: %v", err)
	}

	maskStr := input["mask"].(string)
	img, err := decodeDataURLImage(maskStr)
	if err != nil {
		t.Fatalf("decode inverted mask failed: %v", err)
	}

	bounds := img.Bounds()
	if bounds.Dx() != 4 || bounds.Dy() != 4 {
		t.Fatalf("unexpected mask size: %dx%d", bounds.Dx(), bounds.Dy())
	}

	r, g, b, _ := img.At(0, 0).RGBA()
	if r/257 < 200 || g/257 < 200 || b/257 < 200 {
		t.Fatalf("expected inverted mask area to be near-white, got rgb(%d,%d,%d)", r/257, g/257, b/257)
	}

	r, g, b, _ = img.At(3, 3).RGBA()
	if r/257 > 50 || g/257 > 50 || b/257 > 50 {
		t.Fatalf("expected non-mask area to remain dark, got rgb(%d,%d,%d)", r/257, g/257, b/257)
	}
}

func TestBuildSDXLInpaintingInputDefaultStrength(t *testing.T) {
	req := domain.InpaintFrameRequest{
		ProjectID:   "proj-1",
		FrameID:     "frame-1",
		Prompt:      "fix",
		MaskDataURL: buildTestMaskDataURL(t, 4, 4, func(x, y int) color.Color { return color.Transparent }),
		Strength:    0,
	}

	input, err := BuildSDXLInpaintingInput(req, "data:image/png;base64,frameImage")
	if err != nil {
		t.Fatalf("build input failed: %v", err)
	}
	if input["strength"] != defaultInpaintingStrength {
		t.Fatalf("expected default strength, got %v", input["strength"])
	}
}

func TestBuildSDXLInpaintingInputInvalidMask(t *testing.T) {
	req := domain.InpaintFrameRequest{
		ProjectID:   "proj-1",
		FrameID:     "frame-1",
		Prompt:      "fix",
		MaskDataURL: "not-a-valid-data-url",
		Strength:    0.5,
	}

	_, err := BuildSDXLInpaintingInput(req, "data:image/png;base64,frameImage")
	if err == nil {
		t.Fatalf("expected error for invalid mask data URL")
	}
}

func TestParseSDXLInpaintingOutput(t *testing.T) {
	t.Run("output list", func(t *testing.T) {
		prediction := &Prediction{Output: []any{"https://replicate.delivery/result.png"}}
		url := ParseSDXLInpaintingOutput(prediction)
		if url != "https://replicate.delivery/result.png" {
			t.Fatalf("unexpected output URL: %q", url)
		}
	})

	t.Run("single string", func(t *testing.T) {
		prediction := &Prediction{Output: "https://replicate.delivery/result.png"}
		url := ParseSDXLInpaintingOutput(prediction)
		if url != "https://replicate.delivery/result.png" {
			t.Fatalf("unexpected output URL: %q", url)
		}
	})
}

func buildTestMaskDataURL(t *testing.T, width, height int, pixelFunc func(x, y int) color.Color) string {
	t.Helper()

	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, pixelFunc(x, y))
		}
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("encode test mask: %v", err)
	}

	return encodeDataURL(buf.Bytes(), "image/png")
}

func encodeDataURL(data []byte, mime string) string {
	return "data:" + mime + ";base64," + base64.StdEncoding.EncodeToString(data)
}
