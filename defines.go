package l8f

const (
	VIDEO_FRAME_FORMAT = ".jpg"
	AUDIO_FORMAT       = ".flac"
	OUTPUT_FORMAT      = ".v612"
)

type VideoHeader struct {
	Meta              map[string]string
	VideoUniqueFrames [][]int
	VideoFrames       map[int]int
	AudioSize         int
	VideoFramesSize   int
}

type UniqueFrameDetails struct {
	Hash             string
	FirstFrameNumber int
	Size             int
}

type MakeVideoLumpTemp struct {
	UniqueFrames                []UniqueFrameDetails
	FramesPointerToUniqueFrames map[int]int
}

type MakeVideoLumpTemp2 struct {
	UniqueFrames                []UniqueFrameDetailsNoHash
	FramesPointerToUniqueFrames map[int]int
}

type UniqueFrameDetailsNoHash struct {
	FirstFrameNumber int
	Size             int
	FramePath        string
}
