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

type makeVideoLumpTemp2 struct {
	UniqueFrames                []uniqueFrameDetails
	FramesPointerToUniqueFrames map[int]int
}

type uniqueFrameDetails struct {
	FirstFrameNumber int
	Size             int
	FramePath        string
}
