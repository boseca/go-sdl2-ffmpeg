package sfplay

import (
	"os"
	"testing"
	"time"

	"github.com/veandco/go-sdl2/sdl"
)

func TestMain(m *testing.M) {
	sdl.Main(func() {
		exitcode := m.Run()
		os.Exit(exitcode)
	})
}

func TestSdl(t *testing.T) {
	t.Run("Test Read Video Frame", func(t *testing.T) {
		fileName := "./../examples/player/demo.mp4"

		sdl.Init(sdl.INIT_EVENTS)
		defer sdl.Quit()

		surface := &sdl.Surface{} // dummy surface
		renderer, err := sdl.CreateSoftwareRenderer(surface)
		if err != nil {
			t.Fatalf("Unable to create renderer: %s", err)
			return
		}
		defer renderer.Destroy()

		// register custom SDL event
		frameEvent := sdl.RegisterEvents(1)

		// Start video
		vp := NewVideoPlayDefault(fileName, frameEvent, false)
		if err := vp.Start(renderer); err != nil {
			t.Error("Error starting the video:", err)
			return
		}

		running := true
		success := false
		startTime := time.Now()
		expiry := 10 * time.Second

		// start SDL main loop
		for running {
			for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
				switch event.(type) {
				case *sdl.QuitEvent:
					running = false
				case *sdl.UserEvent:
					// video FrameEvent received
					if event.GetType() == vp.FrameEvent {
						format, access, w, h, err := vp.Texture.Query()
						success = (err == nil && format == 376840196 && access == 1 && w == 1280 && h == 720)
						running = false
					}
				default:
				}
			}
			if time.Since(startTime) > expiry {
				running = false
			} else {
				// read new frame and trigger frameEvent to render the frame
				if _, err := vp.Update(); err != nil {
					running = false
					t.Error("UPDATE error:", err)
				} else {
					sdl.Delay(16)
				}
			}
		}
		if !success {
			t.Error("Custom SDL Event not triggered!")
		}
	})
}
