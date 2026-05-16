package domain

type JobStatus string

const (
	JobStatusQueued    JobStatus = "queued"
	JobStatusRunning   JobStatus = "running"
	JobStatusSucceeded JobStatus = "succeeded"
	JobStatusFailed    JobStatus = "failed"
)

type JobType string

const (
	JobTypeGeneration JobType = "generation"
	JobTypeInpainting JobType = "inpainting"
	JobTypeExport     JobType = "export"
)

type FrameKind string

const (
	FrameKindKey       FrameKind = "key"
	FrameKindGenerated FrameKind = "generated"
	FrameKindInpainted FrameKind = "inpainted"
	FrameKindEdited    FrameKind = "edited"
)

type Frame struct {
	ID           string    `json:"id"`
	ProjectID    string    `json:"projectId"`
	Index        int       `json:"index"`
	ImageURL     string    `json:"imageUrl"`
	ThumbnailURL string    `json:"thumbnailUrl"`
	Kind         FrameKind `json:"kind"`
	Note         string    `json:"note,omitempty"`
	UpdatedAt    string    `json:"updatedAt"`
}

type Project struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

type Job struct {
	ID        string    `json:"id"`
	ProjectID string    `json:"projectId,omitempty"`
	Type      JobType   `json:"type"`
	Status    JobStatus `json:"status"`
	Progress  int       `json:"progress"`
	Message   string    `json:"message"`
	Result    any       `json:"result,omitempty"`
	Error     string    `json:"error,omitempty"`
	CreatedAt string    `json:"createdAt"`
	UpdatedAt string    `json:"updatedAt"`
}

type CreateProjectRequest struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type ProjectResponse struct {
	Project Project `json:"project"`
}

type UpdateProjectRequest struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type GenerateFramesRequest struct {
	ProjectID         string `json:"projectId"`
	Prompt            string `json:"prompt"`
	NegativePrompt    string `json:"negativePrompt,omitempty"`
	FrameCount        int    `json:"frameCount"`
	StartImageDataURL string `json:"startImageDataUrl"`
	EndImageDataURL   string `json:"endImageDataUrl"`
}

type GenerateFramesResult struct {
	Frames []Frame `json:"frames"`
}

type InpaintFrameRequest struct {
	ProjectID   string  `json:"projectId"`
	FrameID     string  `json:"frameId"`
	Prompt      string  `json:"prompt"`
	MaskDataURL string  `json:"maskDataUrl"`
	Strength    float64 `json:"strength"`
}

type InpaintFrameResult struct {
	Frame Frame `json:"frame"`
}

type UpdateFrameRequest struct {
	ProjectID    string `json:"projectId"`
	FrameID      string `json:"frameId"`
	ImageDataURL string `json:"imageDataUrl"`
	Note         string `json:"note,omitempty"`
}

type UpdateFrameResult struct {
	Frame Frame `json:"frame"`
}

type UpdateFrameMetadataRequest struct {
	ProjectID string     `json:"projectId"`
	FrameID   string     `json:"frameId"`
	Kind      *FrameKind `json:"kind,omitempty"`
	Note      *string    `json:"note,omitempty"`
}

type ReorderFramesRequest struct {
	ProjectID string   `json:"projectId"`
	FrameIDs  []string `json:"frameIds"`
}

type ReorderFramesResult struct {
	Frames []Frame `json:"frames"`
}

type ExportVideoRequest struct {
	ProjectID string `json:"projectId"`
	FPS       int    `json:"fps"`
}

type ExportVideoResult struct {
	VideoURL string `json:"videoUrl"`
}
