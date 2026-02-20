package videos612

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

type MakeVideoLumpTemp2 struct {
	UniqueFrames                []UniqueFrameDetailsNoHash
	FramesPointerToUniqueFrames map[int]int
}

type UniqueFrameDetailsNoHash struct {
	FirstFrameNumber int
	Size             int
	FramePath        string
}
