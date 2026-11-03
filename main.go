package main

import (
	"log"
	"os"

	"github.com/egoist/mygo"

	"github.com/IcyBlue17/bil-live-helper/app"
	"github.com/IcyBlue17/bil-live-helper/internal/bili"
	"github.com/IcyBlue17/bil-live-helper/internal/demo"
)

func main() {
	var backend app.Backend
	if os.Getenv("BLH_DEMO") == "1" {
		backend = demo.New()
	} else {
		backend = bili.New()
	}
	app.Bind(backend)

	mygo.App.WhenReady(func() {
		mygo.NewWindow(mygo.WindowOptions{
			Title:           "Bil Live Helper",
			Width:           1180,
			Height:          760,
			MinWidth:        900,
			MinHeight:       600,
			BackgroundColor: "#0e0f13",
			StateKey:        "main",
			URL:             "/",
		})
	})
	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}
