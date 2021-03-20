package main

import (
	"github.com/1lann/dissonance/ffmpeg"
	"github.com/1lann/dissonance/ffplay"
)

func main() {
	stream, err := ffmpeg.NewFFMPEGStreamFromDshow("Microphone (NVIDIA Broadcast)", true)
	if err != nil {
		panic(err)
	}

	out := ffplay.NewFFPlaySink()
	err = out.PlayStream(stream)
	if err != nil {
		panic(err)
	}
}
