package main

import (
	"flag"
	"fmt"
	"os"

	// Export global variables for NVIDIA and AMD drivers so that they use high performance graphics rendering settings. Without this, the default integrated GPU will be used.
	_ "github.com/silbinarywolf/preferdiscretegpu"

	"github.com/boseca/go-sdl2-ffmpeg/sfplay"
	"github.com/veandco/go-sdl2/sdl"
)

var (
	winTitle            string = "Go-SDL2 Render"
	winWidth, winHeight int32  = 800, 600
)

func playVideo(fileName string, audioOn bool, fitWindow bool) int {
	var window *sdl.Window
	var renderer *sdl.Renderer

	// initialize SDL
	if err := sdl.Init(sdl.INIT_VIDEO | sdl.INIT_AUDIO | sdl.INIT_TIMER); err != nil {
		panic(err)
	}
	defer sdl.Quit()

	// register custom SDL event
	frameEvent := sdl.RegisterEvents(1)

	// create Window GUI
	window, err := sdl.CreateWindow(winTitle, sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED, winWidth, winHeight, sdl.WINDOW_SHOWN|sdl.WINDOW_RESIZABLE)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create window: %s\n", err)
		return 1
	}
	defer window.Destroy()

	// create Renderer
	renderer, err = sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create renderer: %s\n", err)
		return 2
	}
	defer renderer.Destroy()

	// // OPTIONAL: sometimes CRASHES the app!
	// surface, err := window.GetSurface()
	// if err != nil {
	// 	panic(err)
	// }
	// surface.FillRect(nil, 0)
	// rect := sdl.Rect{0, 0, 200, 200}
	// surface.FillRect(&rect, 0xffff0000)
	// window.UpdateSurface()

	// Start video
	vp := sfplay.NewVideoPlayDefault(fileName, frameEvent, audioOn)
	vp.Start(renderer)

	// Resize video to fit window
	if fitWindow {
		vp.Resize(winWidth, winHeight)
	}

	running := true

	// start SDL main loop
	for running {
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch t := event.(type) {
			case *sdl.WindowEvent:
				// resize the video
				switch t.Event {
				case sdl.WINDOWEVENT_SIZE_CHANGED:
					if fitWindow {
						vp.Resize(t.Data1, t.Data2)
					}
				}
			case *sdl.KeyboardEvent:
				// close the app when ESC is pressed
				if t.Keysym.Sym == sdl.K_ESCAPE {
					running = false
				}
			case *sdl.QuitEvent:
				running = false
			case *sdl.UserEvent:
				// render video frame texture
				if event.GetType() == vp.FrameEvent {
					renderer.Clear()
					rec := &sdl.Rect{X: 0, Y: 0, W: vp.Width, H: vp.Height}
					renderer.Copy(vp.Texture, rec, rec)
					renderer.Present()
				}
			default:
			}
		}
		// get new frame and trigger frameEvent to render the frame
		if info, err := vp.Update(); err != nil {
			fmt.Println("Update error:", err)
			return 1
		} else {
			window.SetTitle(info)
		}
		sdl.Delay(16)
	}

	return 0
}

func main() {
	// get video file name
	fileName := flag.String("f", "./demo.mp4", "file name")
	audioOn := flag.Bool("a", true, "Audio on/off")
	fitWindow := flag.Bool("fw", true, "Auto-size video to fit window size.")
	flag.Parse()

	// play the video
	os.Exit(playVideo(*fileName, *audioOn, *fitWindow))
}
