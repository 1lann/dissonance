package main

import (
	"github.com/1lann/dissonance/drivers/paudio"
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

	err = pd.PlayStream(rcs)
	if err != nil {
		panic(err)
	}
}
