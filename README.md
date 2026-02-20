# videos612
a slideshow video format. Good for lyrics video.
This format makes use of a framerate of 1. It is uses less CPUs to make a video.
It doesn't support camera video or animations.


## Description of the Format

The general description looks like this:
```
{header_length}
{header}
{audio}
{video_frames_lump}
```

This format uses a framerate of 1

### Description of {header} section

The `{header}` section made up of some subsections. It looks like this

```
meta:
year: 2022
::
video_unique_frames:
{number}: {size}
{number}: {size}
::
video_frames:
{frame_number}: {unique_frame_number}
{frame_number}: {unique_frame_number}
::
binary:
audio: {audio_size_bytes}
video_frames_lump: {laptop_video_size_bytes}
mobile_frames_lump: {mobile_video_size_bytes}
::
```


### Description of the {audio} section

The `{audio}` section takes FLAC data and writes it unparsed to the video format

### Description of the {video_frames_lump} sections

The `{video_frames_lump}` section is made up of a lump file of unique frames.
The unique frames must be of **JPG** format

