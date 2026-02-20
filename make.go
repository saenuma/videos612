package l8f

import (
	"bytes"
	"fmt"
	"image"
	_ "image/jpeg"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

func makeFramesLumpFile(inFramesDirectory, outFilePath string) (MakeVideoLumpTemp2, error) {
	vlt := MakeVideoLumpTemp2{}
	dirFIs, err := os.ReadDir(inFramesDirectory)
	if err != nil {
		return vlt, errors.Wrap(err, "os error")
	}

	inFrameNumbers := make([]int, 0)
	for _, dirFI := range dirFIs {
		if dirFI.IsDir() {
			return vlt, errors.New(fmt.Sprintf("the inFramesDirectory '%s' must not contain any subfolder", inFramesDirectory))
		}
		inFrameNameInt, err := strconv.Atoi(strings.ReplaceAll(dirFI.Name(), VIDEO_FRAME_FORMAT, ""))
		if err != nil {
			return vlt, errors.New(fmt.Sprintf("the file '%s' of inFramesDirectory '%s' is not a number", dirFI.Name(), inFramesDirectory))
		}
		inFrameNumbers = append(inFrameNumbers, inFrameNameInt)
	}

	sort.Ints(inFrameNumbers)

	firstFrameHandle, err := os.Open(filepath.Join(inFramesDirectory, "1"+VIDEO_FRAME_FORMAT))
	if err != nil {
		return vlt, errors.New(fmt.Sprintf("the inFramesDirectory '%s' has no '1%s'", inFramesDirectory, VIDEO_FRAME_FORMAT))
	}
	im, _, err := image.DecodeConfig(firstFrameHandle)
	if err != nil {
		return vlt, errors.Wrap(err, "image error")
	}
	firstWidth := im.Width
	firstHeight := im.Height
	firstFrameHandle.Close()

	// validate same width and height of all the frames
	for _, inFrameNumber := range inFrameNumbers {
		currentFramePath := filepath.Join(inFramesDirectory, fmt.Sprintf("%d%s", inFrameNumber, VIDEO_FRAME_FORMAT))
		currentFrameHandle, err := os.Open(currentFramePath)
		if err != nil {
			return vlt, errors.Wrap(err, "os error")
		}
		currentIm, _, err := image.DecodeConfig(currentFrameHandle)
		if err != nil {
			return vlt, errors.Wrap(err, "image error")
		}
		if currentIm.Width != firstWidth || currentIm.Height != firstHeight {
			return vlt, errors.New(fmt.Sprintf("the width or height of '%s' differs from the first frame", currentFramePath))
		}
	}

	// make temporary lump of unique frames
	outFileHandle, err := os.OpenFile(outFilePath, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		return vlt, errors.Wrap(err, "os error")
	}
	uniqueFrames := make([]UniqueFrameDetailsNoHash, 0) //first frame no and the size
	framesPointer := make(map[int]int)

	checkForUniqueness := func(uniqueFrames []UniqueFrameDetailsNoHash, framePath string) (UniqueFrameDetailsNoHash, error) {
		rawCurrentFrame, _ := os.ReadFile(framePath)

		for _, uniqueFrameDetail := range uniqueFrames {
			rawUniqueFrame, _ := os.ReadFile(uniqueFrameDetail.FramePath)
			retCompare := bytes.Compare(rawCurrentFrame, rawUniqueFrame)
			if retCompare == 0 {
				return uniqueFrameDetail, nil
			}
		}

		return UniqueFrameDetailsNoHash{}, errors.New("frame not found")
	}

	for _, inFrameNumber := range inFrameNumbers {
		currentFramePath := filepath.Join(inFramesDirectory, fmt.Sprintf("%d%s", inFrameNumber, VIDEO_FRAME_FORMAT))

		ufq, err := checkForUniqueness(uniqueFrames, currentFramePath)
		if err == nil {
			framesPointer[inFrameNumber] = ufq.FirstFrameNumber
		} else {
			currentFrameHandle, err := os.Open(currentFramePath)
			if err != nil {
				return vlt, errors.Wrap(err, "os error")
			}
			writtenSize, err := io.Copy(outFileHandle, currentFrameHandle)
			if err != nil {
				return vlt, errors.Wrap(err, "io error")
			}
			uniqueFrames = append(uniqueFrames, UniqueFrameDetailsNoHash{inFrameNumber, int(writtenSize), currentFramePath})
			framesPointer[inFrameNumber] = inFrameNumber
		}
	}
	outFileHandle.Close()

	return MakeVideoLumpTemp2{uniqueFrames, framesPointer}, nil
}

// MakeVideo is good for videos with a lot of stills eg. lyrics videos with a single background.
// the inFramesDirectory must contain png files numbered from 1.png upwards
// the framerate must be stored in the **meta** as a string
func MakeVideo(inVideoFramesPath, inAudioFile string, meta map[string]string, tmpVideoDirectory, outFilePath string) error {
	if !doesPathExists(inVideoFramesPath) {
		return errors.New(fmt.Sprintf("the path '%s' does not exists", inVideoFramesPath))
	}
	if !strings.HasSuffix(inAudioFile, AUDIO_FORMAT) {
		return errors.New(fmt.Sprintf("The inAudioFile must be of type '%s'", AUDIO_FORMAT))
	}
	if !strings.HasSuffix(outFilePath, OUTPUT_FORMAT) {
		return errors.New(fmt.Sprintf("The outFilePath must end with '%s'", OUTPUT_FORMAT))
	}

	for k, v := range meta {
		if strings.Contains(k, "\n") || strings.Contains(v, "\n") {
			return errors.New("The meta elements must not contain newline")
		}
		if strings.Contains(k, ":") || strings.Contains(v, ":") {
			return errors.New("The meta elements must not contain ':' ")
		}
	}

	videoFramesLumpPath := filepath.Join(tmpVideoDirectory, ".tmp_"+untestedRandomString(10)+".bin")

	lvlt, err := makeFramesLumpFile(inVideoFramesPath, videoFramesLumpPath)
	if err != nil {
		panic(err)
	}

	// write meta
	outStr := "meta:\n"
	for metaKey, metaValue := range meta {
		outStr += metaKey + ": " + metaValue + "\n"
	}
	outStr += "::\n"

	// write video_unique_frames
	outStr += "video_unique_frames:\n"
	for _, ufq := range lvlt.UniqueFrames {
		outStr += fmt.Sprintf("%d: %d\n", ufq.FirstFrameNumber, ufq.Size)
	}
	outStr += "::\n"

	// write video_frames info
	outStr += "video_frames:\n"
	for frameNumber, pointedToFrameNumber := range lvlt.FramesPointerToUniqueFrames {
		outStr += fmt.Sprintf("%d: %d\n", frameNumber, pointedToFrameNumber)
	}
	outStr += "::\n"

	// write lumps
	outStr += "binary:\n"
	inAudioFileStat, err := os.Stat(inAudioFile)
	if err != nil {
		return errors.Wrap(err, "os error")
	}

	videoFramesLumpPathStat, err := os.Stat(videoFramesLumpPath)
	if err != nil {
		return errors.Wrap(err, "os error")
	}
	outStr += fmt.Sprintf("audio: %d\n", inAudioFileStat.Size())
	outStr += fmt.Sprintf("video_frames_lump: %d\n", videoFramesLumpPathStat.Size())
	outStr += "::\n"

	outVideoHandle, err := os.OpenFile(outFilePath, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		return errors.Wrap(err, "os error")
	}
	defer outVideoHandle.Close()

	outVideoHandle.WriteString(fmt.Sprintf("%d\n", len(outStr)))
	outVideoHandle.WriteString(outStr)

	inAudioHandle, err := os.Open(inAudioFile)
	if err != nil {
		return errors.Wrap(err, "os error")
	}
	_, err = io.Copy(outVideoHandle, inAudioHandle)
	if err != nil {
		return errors.Wrap(err, "io error")
	}
	videoFramesLumpPathHandle, err := os.Open(videoFramesLumpPath)
	if err != nil {
		return errors.Wrap(err, "os error")
	}
	defer videoFramesLumpPathHandle.Close()
	_, err = io.Copy(outVideoHandle, videoFramesLumpPathHandle)
	if err != nil {
		return errors.Wrap(err, "io error")
	}

	// cleanup
	os.Remove(videoFramesLumpPath)

	return nil
}

// the framerate must be stored in the **meta** as a string
func UpdateMeta(inVideoPath string, meta map[string]string, tmpVideoDirectory, outFilePath string) error {
	for k, v := range meta {
		if strings.Contains(k, "\n") || strings.Contains(v, "\n") {
			return errors.New("The meta elements must not contain newline")
		}
		if strings.Contains(k, ":") || strings.Contains(v, ":") {
			return errors.New("The meta elements must not contain ':' ")
		}
	}
	if !strings.HasSuffix(outFilePath, ".v612") {
		return errors.New("The outFilePath must end with '.v612'")
	}
	vhSize, err := getHeaderLengthFromVideo(inVideoPath)
	if err != nil {
		return err
	}
	vh, err := ReadHeaderFromVideo(inVideoPath)
	if err != nil {
		return err
	}
	vh.Meta = meta
	audioBytes := make([]byte, vh.AudioSize)
	rawVideoHandle, err := os.Open(inVideoPath)
	if err != nil {
		return errors.Wrap(err, "os error")
	}
	defer rawVideoHandle.Close()
	audioOffset := vhSize + 1 + len(fmt.Sprintf("%d", vhSize))
	_, err = rawVideoHandle.ReadAt(audioBytes, int64(audioOffset))
	if err != nil {
		return errors.Wrap(err, "strconv error")
	}
	videoFramesBytes := make([]byte, vh.VideoFramesSize)
	laptopVideoOffset := audioOffset + vh.AudioSize
	_, err = rawVideoHandle.ReadAt(videoFramesBytes, int64(laptopVideoOffset))
	if err != nil {
		return errors.Wrap(err, "strconv error")
	}
	// write meta
	outStr := "meta:\n"
	for metaKey, metaValue := range vh.Meta {
		outStr += metaKey + ": " + metaValue + "\n"
	}
	outStr += "::\n"
	// write unique_frames
	outStr += "video_unique_frames:\n"
	for _, ufq := range vh.VideoUniqueFrames {
		outStr += fmt.Sprintf("%d: %d\n", ufq[0], ufq[1])
	}
	outStr += "::\n"
	// write frames info
	outStr += "video_frames:\n"
	for frameNumber, pointedToFrameNumber := range vh.VideoFrames {
		outStr += fmt.Sprintf("%d: %d\n", frameNumber, pointedToFrameNumber)
	}
	outStr += "::\n"
	// write lumps
	outStr += "binary:\n"
	outStr += fmt.Sprintf("audio: %d\n", vh.AudioSize)
	outStr += fmt.Sprintf("video_frames_lump: %d\n", vh.VideoFramesSize)
	outStr += "::\n"
	outVideoHandle, err := os.OpenFile(outFilePath, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		return errors.Wrap(err, "os error")
	}
	defer outVideoHandle.Close()
	outVideoHandle.WriteString(fmt.Sprintf("%d\n", len(outStr)))
	outVideoHandle.WriteString(outStr)
	outVideoHandle.Write(audioBytes)
	outVideoHandle.Write(videoFramesBytes)
	return nil
}
