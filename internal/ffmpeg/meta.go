package ffmpeg

import (
	"foxy/internal/vision"
	"math"
	"strconv"
	"strings"
)

type StreamMeta struct {
	Index            *int    `json:"index,omitempty"`
	CodecName        *string `json:"codec_name,omitempty"`
	CodecLongName    *string `json:"codec_long_name,omitempty"`
	Profile          *string `json:"profile,omitempty"`
	CodecType        *string `json:"codec_type,omitempty"`
	CodecTagString   *string `json:"codec_tag_string,omitempty"`
	CodecTag         *string `json:"codec_tag,omitempty"`
	Width            *int    `json:"width,omitempty"`
	Height           *int    `json:"height,omitempty"`
	CodedWidth       *int    `json:"coded_width,omitempty"`
	CodedHeight      *int    `json:"coded_height,omitempty"`
	ClosedCaptions   *int    `json:"closed_captions,omitempty"`
	FilmGrain        *int    `json:"film_grain,omitempty"`
	HasBFrames       *int    `json:"has_b_frames,omitempty"`
	PixFmt           *string `json:"pix_fmt,omitempty"`
	Level            *int    `json:"level,omitempty"`
	ColorRange       *string `json:"color_range,omitempty"`
	ColorSpace       *string `json:"color_space,omitempty"`
	ColorTransfer    *string `json:"color_transfer,omitempty"`
	ColorPrimaries   *string `json:"color_primaries,omitempty"`
	ChromaLocation   *string `json:"chroma_location,omitempty"`
	FieldOrder       *string `json:"field_order,omitempty"`
	Refs             *int    `json:"refs,omitempty"`
	IsAvc            *string `json:"is_avc,omitempty"`
	NalLengthSize    *string `json:"nal_length_size,omitempty"`
	Id               *string `json:"id,omitempty"`
	RFrameRate       *string `json:"r_frame_rate,omitempty"`
	AvgFrameRate     *string `json:"avg_frame_rate,omitempty"`
	TimeBase         *string `json:"time_base,omitempty"`
	StartPts         *int    `json:"start_pts,omitempty"`
	StartTime        *string `json:"start_time,omitempty"`
	DurationTs       *int    `json:"duration_ts,omitempty"`
	Duration         *string `json:"duration,omitempty"`
	BitRate          *string `json:"bit_rate,omitempty"`
	BitsPerRawSample *string `json:"bits_per_raw_sample,omitempty"`
	NbFrames         *string `json:"nb_frames,omitempty"`
	ExtradataSize    *int    `json:"extradata_size,omitempty"`
	SampleFmt        *string `json:"sample_fmt,omitempty"`
	SampleRate       *string `json:"sample_rate,omitempty"`
	Channels         *int    `json:"channels,omitempty"`
	ChannelLayout    *string `json:"channel_layout,omitempty"`
	BitsPerSample    *int    `json:"bits_per_sample,omitempty"`
	InitialPadding   *int    `json:"initial_padding,omitempty"`
}

type FormatMeta struct {
	Filename       *string `json:"filename,omitempty"`
	NbStreams      *int    `json:"nb_streams,omitempty"`
	NbPrograms     *int    `json:"nb_programs,omitempty"`
	NbStreamGroups *int    `json:"nb_stream_groups,omitempty"`
	FormatName     *string `json:"format_name,omitempty"`
	FormatLongName *string `json:"format_long_name,omitempty"`
	StartTime      *string `json:"start_time,omitempty"`
	Duration       *string `json:"duration,omitempty"`
	Size           *string `json:"size,omitempty"`
	BitRate        *string `json:"bit_rate,omitempty"`
	ProbeScore     *int    `json:"probe_score,omitempty"`
}

type Meta struct {
	Streams   []StreamMeta `json:"streams"`
	Format    *FormatMeta  `json:"format,omitempty"`
	Keyframes []string     `json:"keyframes,omitempty"`
}

func (meta *Meta) GetVideoMetadata() *vision.VideoMetadata {
	var videoTrack *StreamMeta
	for _, stream := range meta.Streams {
		if stream.CodecType != nil && *stream.CodecType == "video" {
			videoTrack = &stream
			break
		}
	}

	if videoTrack == nil || videoTrack.Width == nil || videoTrack.Height == nil || videoTrack.AvgFrameRate == nil {
		return nil
	}

	var d float64
	var err error
	if videoTrack.Duration == nil {
		if meta.Format.Duration == nil {
			return nil
		}

		d, err = strconv.ParseFloat(*meta.Format.Duration, 64)
		if err != nil {
			return nil
		}
	} else {
		d, err = strconv.ParseFloat(*videoTrack.Duration, 64)
		if err != nil {
			return nil
		}
	}

	durParts := strings.Split(*videoTrack.AvgFrameRate, "/")
	if len(durParts) != 2 {
		return nil
	}

	t1, err := strconv.ParseFloat(durParts[0], 64)
	if err != nil {
		return nil
	}

	t2, err := strconv.ParseFloat(durParts[1], 64)
	if err != nil {
		return nil
	}

	var fc int
	if videoTrack.NbFrames != nil {
		fc, err = strconv.Atoi(*videoTrack.NbFrames)
		if err != nil {
			return nil
		}
	} else {
		fc = int(math.Floor(d * (t1 / t2)))
	}

	return &vision.VideoMetadata{
		Width:         *videoTrack.Width,
		Height:        *videoTrack.Height,
		Duration:      d,
		FPS:           t1 / t2,
		FrameCount:    fc,
		KeyframeCount: len(meta.Keyframes),
	}
}
