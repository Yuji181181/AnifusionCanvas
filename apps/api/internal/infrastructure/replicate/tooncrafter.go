package replicate

import "github.com/haseg/anifusion-canvas/apps/api/internal/domain"

// ToonCrafterMaxWidth is the maximum output width for ToonCrafter.
const ToonCrafterMaxWidth = 640

// ToonCrafterMaxHeight is the maximum output height for ToonCrafter.
const ToonCrafterMaxHeight = 360

// BuildToonCrafterInput converts a GenerateFramesRequest into the input map
// expected by the Replicate ToonCrafter model.
func BuildToonCrafterInput(req domain.GenerateFramesRequest) map[string]any {
	input := map[string]any{
		"image_1":    req.StartImageDataURL,
		"image_2":    req.EndImageDataURL,
		"prompt":     req.Prompt,
		"max_width":  ToonCrafterMaxWidth,
		"max_height": ToonCrafterMaxHeight,
	}

	if req.NegativePrompt != "" {
		input["negative_prompt"] = req.NegativePrompt
	}

	return input
}

// ParseToonCrafterOutput extracts the MP4 video URL from a completed
// ToonCrafter prediction.
func ParseToonCrafterOutput(prediction *Prediction) string {
	return prediction.OutputURL()
}
