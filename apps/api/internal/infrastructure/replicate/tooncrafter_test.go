package replicate

import (
	"testing"

	"github.com/haseg/anifusion-canvas/apps/api/internal/domain"
)

func TestBuildToonCrafterInput(t *testing.T) {
	req := domain.GenerateFramesRequest{
		ProjectID:         "proj-1",
		Prompt:            "a character turning around",
		NegativePrompt:    "blurry, low quality",
		FrameCount:        3,
		StartImageDataURL: "data:image/png;base64,start",
		EndImageDataURL:   "data:image/png;base64,end",
	}

	input := BuildToonCrafterInput(req)

	if input["image_1"] != "data:image/png;base64,start" {
		t.Fatalf("unexpected image_1: %v", input["image_1"])
	}
	if input["image_2"] != "data:image/png;base64,end" {
		t.Fatalf("unexpected image_2: %v", input["image_2"])
	}
	if input["prompt"] != "a character turning around" {
		t.Fatalf("unexpected prompt: %v", input["prompt"])
	}
	if input["negative_prompt"] != "blurry, low quality" {
		t.Fatalf("unexpected negative_prompt: %v", input["negative_prompt"])
	}
	if input["max_width"] != ToonCrafterMaxWidth {
		t.Fatalf("unexpected max_width: %v", input["max_width"])
	}
	if input["max_height"] != ToonCrafterMaxHeight {
		t.Fatalf("unexpected max_height: %v", input["max_height"])
	}
}

func TestBuildToonCrafterInputWithoutNegativePrompt(t *testing.T) {
	req := domain.GenerateFramesRequest{
		Prompt:            "walk cycle",
		StartImageDataURL: "data:image/png;base64,start",
		EndImageDataURL:   "data:image/png;base64,end",
	}

	input := BuildToonCrafterInput(req)

	if _, ok := input["negative_prompt"]; ok {
		t.Fatalf("expected negative_prompt to be absent when empty")
	}
}

func TestParseToonCrafterOutput(t *testing.T) {
	prediction := &Prediction{Output: "https://replicate.delivery/test/output.mp4"}
	url := ParseToonCrafterOutput(prediction)
	if url != "https://replicate.delivery/test/output.mp4" {
		t.Fatalf("unexpected output URL: %q", url)
	}
}

func TestParseToonCrafterOutputNonString(t *testing.T) {
	prediction := &Prediction{Output: 42}
	url := ParseToonCrafterOutput(prediction)
	if url != "" {
		t.Fatalf("expected empty output URL for non-string, got %q", url)
	}
}
