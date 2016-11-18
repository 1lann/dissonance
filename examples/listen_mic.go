package main

import (
	"github.com/1lann/dissonance/drivers/paudio"
	"github.com/1lann/dissonance/filters/vad"
)

func main() {
	pd, err := paudio.NewPlaybackDevice()
	if err != nil {
		panic(err)
	}

	rc, err := paudio.NewRecordingDevice()
	if err != nil {
		panic(err)
	}

	rcs, err := rc.OpenStream()
	if err != nil {
		panic(err)
	}

	filter := vad.NewFilter(0.1)
	str := filter.Filter(rcs)

	err = pd.PlayStream(str)
	if err != nil {
		panic(err)
	}

	pd.Close()
}
