package main

import (
	"log"
	"os"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/plugins/updater"

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
	// The configured name, BilLiveHelper, is the executable's; this is the
	// one the app shows and keeps its data under.
	mygo.App.SetName("Bil Live Helper")
	app.Bind(backend)
	// Checks GitHub releases 10s after launch and then daily; the window is
	// in Chinese like the rest of the app.
	mygo.Use(updater.New(updater.Options{Language: "zh-Hans"}))

	mygo.App.WhenReady(func() {
		updater.OnChange(func() { _ = app.UpdateChanged.Broadcast(app.CurrentUpdateStatus()) })
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
