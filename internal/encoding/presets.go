package encoding

type HLSPreset struct {
	SegmentSeconds int    // hls time
	GOP            int    //keyframse interval (frames)
	VideoCodec     string // libx264
	AudioCodec     string // aac
	Preset         string // veryfast
}

func DefaultHLSPreset() HLSPreset {
	return HLSPreset{
		SegmentSeconds: 4,
		GOP:            48,
		VideoCodec:     "libx264",
		AudioCodec:     "aac",
		Preset:         "veryfast",
	}
}
