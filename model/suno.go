package models

import "time"

type TaskStatus string

const (
	TaskStatusNotStart   TaskStatus = "NOT_START"
	TaskStatusSubmitted  TaskStatus = "SUBMITTED"
	TaskStatusQueued     TaskStatus = "QUEUED"
	TaskStatusInProgress TaskStatus = "IN_PROGRESS"
	TaskStatusFailure    TaskStatus = "FAILURE"
	TaskStatusSuccess    TaskStatus = "SUCCESS"
	TaskStatusUnknown    TaskStatus = "UNKNOWN"
)

func (status TaskStatus) IsDone() bool {
	return status == TaskStatusSuccess ||
		status == TaskStatusFailure ||
		status == TaskStatusUnknown
}

type SubmitGenSongReq struct {
	Prompt               string `json:"prompt"`
	Mv                   string `json:"mv"`
	Title                string `json:"title"`
	Tags                 string `json:"tags"`
	GptDescriptionPrompt string `json:"gpt_description_prompt,omitempty"`

	TaskID           string   `json:"task_id"`
	ContinueAt       *float64 `json:"continue_at,omitempty"`
	ContinueClipId   *string  `json:"continue_clip_id,omitempty"`
	MakeInstrumental bool     `json:"make_instrumental"`
}

// Response 是整个 JSON 响应的结构体
type ResponseMusic struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Data    string `json:"data"`
}

type ResponseFeed struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Data    Data   `json:"data"`
}

// Data 包含任务和音频数据的结构体
type Data struct {
	TaskID      string       `json:"task_id"`
	NotifyHook  string       `json:"notify_hook"`
	Action      string       `json:"action"`
	Status      TaskStatus   `json:"status"`
	FailReason  string       `json:"fail_reason"`
	SubmitTime  int64        `json:"submit_time"`
	StartTime   int64        `json:"start_time"`
	FinishTime  int64        `json:"finish_time"`
	Progress    string       `json:"progress"`
	AudioTracks []AudioTrack `json:"data"`
}

// AudioTrack 代表单个音频条目的结构体
type AudioTrack struct {
	ID                string      `json:"id"`
	Title             string      `json:"title"`
	Handle            string      `json:"handle"`
	Status            string      `json:"status"`
	UserID            string      `json:"user_id"`
	IsLiked           bool        `json:"is_liked"`
	Metadata          Metadata    `json:"metadata"`
	Reaction          interface{} `json:"reaction"` // 假设 reaction 可能是一个复杂结构，这里暂时用 interface{} 替代
	AudioURL          string      `json:"audio_url"`
	ImageURL          string      `json:"image_url"`
	IsPublic          bool        `json:"is_public"`
	VideoURL          string      `json:"video_url"`
	CreatedAt         time.Time   `json:"created_at"`
	IsTrashed         bool        `json:"is_trashed"`
	ModelName         string      `json:"model_name"`
	PlayCount         int         `json:"play_count"`
	DisplayName       string      `json:"display_name"`
	UpvoteCount       int         `json:"upvote_count"`
	ImageLargeURL     string      `json:"image_large_url"`
	IsVideoPending    bool        `json:"is_video_pending"`
	IsHandleUpdated   bool        `json:"is_handle_updated"`
	MajorModelVersion string      `json:"major_model_version"`
}

// Metadata 代表音频条目的元数据结构体
type Metadata struct {
	Tags                 string      `json:"tags"`
	Type                 string      `json:"type"`
	Prompt               string      `json:"prompt"`
	Stream               bool        `json:"stream"`
	History              interface{} `json:"history"` // 假设 history 可能是一个复杂结构，这里暂时用 interface{} 替代
	Duration             float64     `json:"duration"`
	ErrorType            interface{} `json:"error_type"`     // 假设 error_type 可能是一个复杂结构，这里暂时用 interface{} 替代
	ErrorMessage         interface{} `json:"error_message"`  // 假设 error_message 可能是一个复杂结构，这里暂时用 interface{} 替代
	ConcatHistory        interface{} `json:"concat_history"` // 假设 concat_history 可能是一个复杂结构，这里暂时用 interface{} 替代
	RefundCredits        bool        `json:"refund_credits"`
	AudioPromptID        interface{} `json:"audio_prompt_id"`        // 假设 audio_prompt_id 可能是一个复杂结构，这里暂时用 interface{} 替代
	GptDescriptionPrompt interface{} `json:"gpt_description_prompt"` // 假设 gpt_description_prompt 可能是一个复杂结构，这里暂时用 interface{} 替代
}
