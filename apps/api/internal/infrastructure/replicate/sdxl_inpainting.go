package replicate

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"strings"

	"github.com/haseg/anifusion-canvas/apps/api/internal/domain"
)

const (
	defaultInpaintingStrength = 0.75
	defaultInpaintingSteps    = 25
)

// BuildSDXLInpaintingInput converts an InpaintFrameRequest into the input map
// expected by the Replicate SDXL Inpainting model.
//
// IMPORTANT: The frontend draws a black mask on the areas to fix, but SDXL
// Inpainting expects a WHITE mask for areas to inpaint. This function inverts
// the mask colors before sending to Replicate.
func BuildSDXLInpaintingInput(req domain.InpaintFrameRequest, frameImageURL string) (map[string]any, error) {
	invertedMask, err := decodeDataURLImage(req.MaskDataURL)
	if err != nil {
		return nil, fmt.Errorf("decode mask: %w", err)
	}
	invertImage(invertedMask)

	invertedMaskDataURL, err := encodeDataURLImage(invertedMask)
	if err != nil {
		return nil, fmt.Errorf("encode inverted mask: %w", err)
	}

	strength := req.Strength
	if strength <= 0 {
		strength = defaultInpaintingStrength
	}

	input := map[string]any{
		"image":               frameImageURL,
		"mask":                invertedMaskDataURL,
		"prompt":              req.Prompt,
		"strength":            strength,
		"num_inference_steps": defaultInpaintingSteps,
	}

	return input, nil
}

// ParseSDXLInpaintingOutput extracts the result image URL from a completed
// SDXL Inpainting prediction. Returns the first URL from the output list.
func ParseSDXLInpaintingOutput(prediction *Prediction) string {
	urls := prediction.OutputURLList()
	if len(urls) > 0 {
		return urls[0]
	}
	return prediction.OutputURL()
}

// decodeDataURLImage decodes a PNG data URL into an image.Image.
func decodeDataURLImage(dataURL string) (image.Image, error) {
	data, err := decodeDataURL(dataURL)
	if err != nil {
		return nil, err
	}
	img, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("decode PNG: %w", err)
	}
	return img, nil
}

// invertImage inverts every pixel in the image (white↔black, color inversion).
func invertImage(img image.Image) {
	bounds := img.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, a := img.At(x, y).RGBA()
			rr := uint8(255 - r/257)
			gg := uint8(255 - g/257)
			bb := uint8(255 - b/257)
			aa := uint8(a / 257)
			if rgba, ok := img.(*image.RGBA); ok {
				rgba.SetRGBA(x, y, color.RGBA{R: rr, G: gg, B: bb, A: aa})
			} else if nrgba, ok := img.(*image.NRGBA); ok {
				nrgba.SetNRGBA(x, y, color.NRGBA{R: rr, G: gg, B: bb, A: aa})
			}
		}
	}
}

// encodeDataURLImage encodes an image.Image as a PNG data URL.
func encodeDataURLImage(img image.Image) (string, error) {
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return "", fmt.Errorf("encode PNG: %w", err)
	}
	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}

func decodeDataURL(dataURL string) ([]byte, error) {
	idx := strings.Index(dataURL, "base64,")
	if idx < 0 {
		idx = strings.Index(dataURL, ",")
		if idx < 0 {
			return nil, fmt.Errorf("invalid data URL: no comma found")
		}
		return []byte(dataURL[idx+1:]), nil
	}
	return base64.StdEncoding.DecodeString(dataURL[idx+7:])
}
