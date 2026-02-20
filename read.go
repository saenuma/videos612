package l8f

import (
	"bytes"
	"fmt"
	"image"
	_ "image/png"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

func getHeaderLengthFromVideo(inVideoPath string) (int, error) {
	if !doesPathExists(inVideoPath) {
		return 0, errors.New(fmt.Sprintf("the path '%s' does not exists", inVideoPath))
	}
	if !strings.HasSuffix(inVideoPath, OUTPUT_FORMAT) {
		return 0, errors.New(fmt.Sprintf("The inVideoPath must be of type '%s'", OUTPUT_FORMAT))
	}

	rawVideoHandle, err := os.Open(inVideoPath)
	if err != nil {
		return 0, errors.Wrap(err, "os error")
	}
	defer rawVideoHandle.Close()

	var count int64
	var headerLengthStr string
	for {
		inByte := make([]byte, 1)
		_, err := rawVideoHandle.ReadAt(inByte, count)
		if err != nil {
			return 0, errors.Wrap(err, "os error")
		}
		if string(inByte) != "\n" {
			headerLengthStr += string(inByte)
			count += 1
		} else {
			break
		}
		continue
	}

	headerLength, err := strconv.Atoi(headerLengthStr)
	if err != nil {
		return 0, errors.Wrap(err, "strconv error")
	}

	return headerLength, nil
}

func ReadHeaderFromVideo(inVideoPath string) (VideoHeader, error) {
	evh := VideoHeader{}
	if !doesPathExists(inVideoPath) {
		return evh, errors.New(fmt.Sprintf("the path '%s' does not exists", inVideoPath))
	}
	if !strings.HasSuffix(inVideoPath, OUTPUT_FORMAT) {
		return evh, errors.New(fmt.Sprintf("The inVideoPath must be of type '%s'", OUTPUT_FORMAT))
	}

	rawVideoHandle, err := os.Open(inVideoPath)
	if err != nil {
		return evh, errors.Wrap(err, "os error")
	}
	defer rawVideoHandle.Close()

	headerLength, err := getHeaderLengthFromVideo(inVideoPath)
	if err != nil {
		return evh, errors.Wrap(err, "strconv error")
	}
	headerBytes := make([]byte, headerLength)
	readBegin := int64(len(strconv.Itoa(headerLength))) + 1
	_, err = rawVideoHandle.ReadAt(headerBytes, readBegin)
	if err != nil {
		return evh, errors.Wrap(err, "os error")
	}
	headerStr := string(headerBytes)

	// begin parsing Video header
	metaBeginPart := strings.Index(headerStr, "meta:")
	if metaBeginPart != -1 {
		metaEndPart := strings.Index(headerStr[metaBeginPart:], "::")
		if metaEndPart == -1 {
			return evh, errors.New("Bad Header: meta section must end with a '::'.")
		}
		metaPart := headerStr[metaBeginPart+len("meta:\n") : metaBeginPart+metaEndPart]
		meta := make(map[string]string)
		for _, line := range strings.Split(metaPart, "\n") {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			partsOfLine := strings.Split(line, ":")
			meta[partsOfLine[0]] = strings.TrimSpace(partsOfLine[1])
		}
		evh.Meta = meta
	}

	vUniqueFramesBeginPart := strings.Index(headerStr, "video_unique_frames:")
	vUniqueFramesEndPart := strings.Index(headerStr[vUniqueFramesBeginPart:], "::")
	if vUniqueFramesEndPart == -1 {
		return evh, errors.New("Bad Header: video_unique_frames section must end with a '::'.")
	}
	vUniqueFramesPart := headerStr[vUniqueFramesBeginPart+len("video_unique_frames:\n") : vUniqueFramesBeginPart+vUniqueFramesEndPart]
	vUniqueFrames := make([][]int, 0)
	for _, line := range strings.Split(vUniqueFramesPart, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		partsOfLine := strings.Split(line, ":")
		f1, err := strconv.Atoi(strings.TrimSpace(partsOfLine[0]))
		if err != nil {
			return evh, errors.Wrap(err, "strconv error")
		}
		f2, err := strconv.Atoi(strings.TrimSpace(partsOfLine[1]))
		if err != nil {
			return evh, errors.Wrap(err, "strconv error")
		}

		vUniqueFrames = append(vUniqueFrames, []int{f1, f2})
	}
	evh.VideoUniqueFrames = vUniqueFrames

	vFramesBeginPart := strings.LastIndex(headerStr, "video_frames:")
	vFramesEndPart := strings.Index(headerStr[vFramesBeginPart:], "::")
	if vFramesEndPart == -1 {
		return evh, errors.New("Bad Header: video_frames section must end with a '::'.")
	}
	vFramesPart := headerStr[vFramesBeginPart+len("video_frames:\n") : vFramesEndPart+vFramesBeginPart]
	vFrames := make(map[int]int)
	for _, line := range strings.Split(vFramesPart, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		partsOfLine := strings.Split(line, ":")
		frame1Int, err := strconv.Atoi(partsOfLine[0])
		if err != nil {
			return evh, errors.Wrap(err, "strconv error")
		}
		frame2Int, err := strconv.Atoi(strings.TrimSpace(partsOfLine[1]))
		if err != nil {
			return evh, errors.Wrap(err, "strconv error")
		}
		vFrames[frame1Int] = frame2Int
	}
	evh.VideoFrames = vFrames

	binaryBeginPart := strings.Index(headerStr, "binary:")
	binaryEndPart := strings.Index(headerStr[binaryBeginPart:], "::")
	if binaryBeginPart == -1 {
		return evh, errors.New("Bad Header: chapters section be present.")
	}
	binaryPart := headerStr[binaryBeginPart+len("binary:\n") : binaryBeginPart+binaryEndPart]
	partsOfBinaryPart := strings.Split(binaryPart, "\n")
	audioPart := partsOfBinaryPart[0]
	videoFramesPart := partsOfBinaryPart[1]

	audioSizeStr := audioPart[len("audio: "):]
	audioSizeInt, err := strconv.Atoi(audioSizeStr)
	if err != nil {
		return evh, errors.Wrap(err, "strconv error")
	}
	videoFramesSizeStr := videoFramesPart[len("video_frames_lump: "):]
	videoFramesSizeInt, err := strconv.Atoi(videoFramesSizeStr)
	if err != nil {
		return evh, errors.Wrap(err, "strconv error")
	}

	evh.AudioSize = audioSizeInt
	evh.VideoFramesSize = videoFramesSizeInt

	return evh, nil
}

// The audio is []bytes but it should contain 'mp3' audio
func ReadAudio(inVideoPath string) ([]byte, error) {
	vhSize, err := getHeaderLengthFromVideo(inVideoPath)
	if err != nil {
		return nil, err
	}
	vh, err := ReadHeaderFromVideo(inVideoPath)
	if err != nil {
		return nil, err
	}

	audioBytes := make([]byte, vh.AudioSize)

	rawVideoHandle, err := os.Open(inVideoPath)
	if err != nil {
		return nil, errors.Wrap(err, "os error")
	}
	defer rawVideoHandle.Close()

	audioOffset := vhSize + 1 + len(fmt.Sprintf("%d", vhSize))
	_, err = rawVideoHandle.ReadAt(audioBytes, int64(audioOffset))
	if err != nil {
		return nil, errors.Wrap(err, "strconv error")
	}

	return audioBytes, nil
}

// Read frame for 1 seconds, starting from the 'seconds' parameter
// 'seconds' parameter starts from 1
func ReadLaptopFrame(inVideoPath string, seconds int) (*image.Image, error) {
	vhSize, err := getHeaderLengthFromVideo(inVideoPath)
	if err != nil {
		return nil, err
	}

	vh, err := ReadHeaderFromVideo(inVideoPath)
	if err != nil {
		return nil, err
	}

	rawVideoHandle, err := os.Open(inVideoPath)
	if err != nil {
		return nil, errors.Wrap(err, "os error")
	}
	defer rawVideoHandle.Close()

	audioOffset := vhSize + 1 + len(fmt.Sprintf("%d", vhSize))
	videoBytes := make([]byte, vh.VideoFramesSize)
	videoOffset := audioOffset + vh.AudioSize
	_, err = rawVideoHandle.ReadAt(videoBytes, int64(videoOffset))
	if err != nil {
		return nil, errors.Wrap(err, "strconv error")
	}

	allFrames := make([]int, 0)
	for k := range vh.VideoFrames {
		allFrames = append(allFrames, k)
	}

	sort.Ints(allFrames)

	pointedToFrameNumber := vh.VideoFrames[seconds]

	// unpack the right frame
	readFrameOffset := 0
	toReadSize := 0

	for _, uniqueFrameDetails := range vh.VideoUniqueFrames {
		if uniqueFrameDetails[0] == pointedToFrameNumber {
			toReadSize = int(uniqueFrameDetails[1])
			break
		} else {
			readFrameOffset += int(uniqueFrameDetails[1])
		}
	}

	currentFrameBytes := videoBytes[readFrameOffset : readFrameOffset+toReadSize]
	img, _, err := image.Decode(bytes.NewReader(currentFrameBytes))
	if err != nil {
		return nil, errors.Wrap(err, "image error")
	}

	return &img, nil
}

// This checks the length of the video using the frames itself
// It doesn't check against the audio data embedded in it
func GetVideoLength(inVideoPath string) (int, error) {
	if !doesPathExists(inVideoPath) {
		return 0, errors.New(fmt.Sprintf("the path '%s' does not exists", inVideoPath))
	}
	if !strings.HasSuffix(inVideoPath, OUTPUT_FORMAT) {
		return 0, errors.New(fmt.Sprintf("The inVideoPath must be of type '%s'", OUTPUT_FORMAT))
	}

	vh, err := ReadHeaderFromVideo(inVideoPath)
	if err != nil {
		return 0, err
	}

	return len(vh.VideoFrames), nil
}
