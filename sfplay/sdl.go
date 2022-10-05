package sfplay

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"image"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/speaker"
	"github.com/veandco/go-sdl2/sdl"
)

// VideoPlay contains Video information, channels for reading the video frames and texture for rendering the video frames
type VideoPlay struct {
	Filename          string          // = "./demo.mp4"
	Width             int32           // = 1280
	Height            int32           // = 720
	FrameBufferSize   int32           // = 1024
	SampleRate        beep.SampleRate // = 44100
	ChannelCount      int32           // = 2
	BitDepth          int32           // = 8
	SampleBufferSize  int32           // = 32 * channelCount * bitDepth * 1024
	SpeakerSampleRate beep.SampleRate // = 44100
	AudioOn           bool            // = true // for WSL2 on Windows 10, set it to `false`` to disable the audio because three is no sound driver. Win 11 should not have this issue.
	HwAccelFlags      int32           // Hardware Acceleration Flags (default 1 - Auto)
	FrameEvent        uint32
	Texture           *sdl.Texture

	chTicker      <-chan time.Time
	chErrs        <-chan error
	chFrameBuffer <-chan *image.RGBA
	fps           int
	frameCount    int64
	framesPlayed  int
	playbackFPS   int
	chPerSecond   <-chan time.Time
	last          time.Time
	deltaTime     float64

	codecName     string
	codecLongName string
	bitRate       int64
}

// NewVideoPlayDefault creates a new VideoPlay structure with default values
func NewVideoPlayDefault(fileName string, frameEvent uint32, audioOn bool) *VideoPlay {
	vp := new(VideoPlay)

	vp.Filename = fileName
	vp.Width = 1280
	vp.Height = 720
	vp.FrameBufferSize = 1024
	vp.SampleRate = 44100
	vp.ChannelCount = 2
	vp.BitDepth = 8
	vp.SampleBufferSize = 32 * vp.ChannelCount * vp.BitDepth * 1024
	vp.SpeakerSampleRate = 44100
	vp.AudioOn = audioOn
	vp.FrameEvent = frameEvent

	return vp
}

// Start begins reading samples and frames of the media file
func (vp *VideoPlay) Start(render *sdl.Renderer) error {
	var err error

	// Initialize the audio speaker.
	if vp.AudioOn {
		err = speaker.Init(vp.SampleRate,
			vp.SpeakerSampleRate.N(time.Second/10))
		if err != nil {
			return err
		}
	}

	// Open the media file.
	media, err := NewMedia(vp.Filename)
	if err != nil {
		return err
	}

	// Get the FPS for playing video frames.
	videoFPS, _ := media.Streams()[0].FrameRate()
	vp.frameCount = media.Streams()[0].FrameCount()
	vp.codecName = media.Streams()[0].CodecName()
	vp.codecLongName = media.Streams()[0].CodecLongName()
	vp.bitRate = media.Streams()[0].BitRate()
	vp.Width = int32(media.streams[0].innerStream().codec.width)
	vp.Height = int32(media.streams[0].innerStream().codec.height)
	vp.HwAccelFlags = int32(media.streams[0].innerStream().codec.hwaccel_flags)

	// SPF for frame ticker.
	spf := 1.0 / float64(videoFPS)
	frameDuration, err := time.ParseDuration(fmt.Sprintf("%fs", spf))
	if err != nil {
		return err
	}

	// Create Texture for drawing video frames.
	vp.Texture, err = render.CreateTexture(uint32(sdl.PIXELFORMAT_RGBA32), sdl.TEXTUREACCESS_STREAMING, vp.Width, vp.Height) // pixes must be in Little-Endian order
	// vp.Texture, err = render.CreateTexture(uint32(sdl.PIXELFORMAT_RGBA8888), sdl.TEXTUREACCESS_STREAMING, vp.Width, vp.Height) // pixes must be in Big-Endian order
	// vp.Texture, err = render.CreateTexture(sdl.PIXELFORMAT_IYUV, sdl.TEXTUREACCESS_STREAMING, vp.Width, vp.Height)
	if err != nil {
		return err
	}

	// Start decoding streams.
	var sampleSource <-chan [2]float64
	vp.chFrameBuffer, sampleSource, vp.chErrs, err = vp.ReadVideoAndAudio(media)
	if err != nil {
		return err
	}

	// Start playing audio samples.
	if vp.AudioOn {
		speaker.Play(streamSamples(sampleSource))
	}

	vp.chTicker = time.Tick(frameDuration)

	// Setup metrics.
	vp.last = time.Now()
	vp.fps = 0
	vp.chPerSecond = time.Tick(time.Second)
	vp.framesPlayed = 0
	vp.playbackFPS = 0

	return nil
}

// Update handles the channels information for video frames, errors and window title information.
// This has to be called from the SDL main loop to be in same thread to be able to update SDL.
// If this is called from a different thread the SDL will crash with `signal SIGSEGV: segmentation violation code`
func (vp *VideoPlay) Update() (string, error) {
	info := ""
	// Compute delta time
	vp.deltaTime = time.Since(vp.last).Seconds()
	vp.last = time.Now()

	// Check for incoming errors.
	select {
	case err, ok := <-vp.chErrs:
		if ok {
			return info, err
		}
	default:
	}

	// Read video frames and trigger draw event.		(Frame size 3686400 = 1280*720*4)
	select {
	case <-vp.chTicker:
		frame, ok := <-vp.chFrameBuffer
		if ok {
			// NOTE: The pixel data must be in the pixel format of the Texture
			pixels := aUint8to32(frame.Pix)
			vp.Texture.UpdateRGBA(nil, pixels, frame.Stride/4) // byte is alias for uint8
			vp.framesPlayed++
			vp.playbackFPS++

			// send to SDL render
			_, err := sdl.PushEvent(&sdl.UserEvent{
				Type:      vp.FrameEvent,
				Timestamp: uint32(time.Now().Unix()),
			})
			if err != nil {
				return info, err
			}
		}
	default:
	}

	vp.fps++

	// Update metrics in the window title.
	select {
	case <-vp.chPerSecond:
		info = fmt.Sprintf("%s | FPS: %d | dt: %f | Frames: %d | Video FPS: %d", "Video", vp.fps, vp.deltaTime, vp.framesPlayed, vp.playbackFPS)
		vp.fps = 0
		vp.playbackFPS = 0

	default:
	}

	return info, nil
}

// Resize changes the SDL Texture size used for rendering the video frame
func (vp *VideoPlay) Resize(w int32, h int32) {
	vp.Width = w
	vp.Height = h
}

// Helper functions -------------------------

// ReadVideoAndAudio reads video and audio frames
// from the opened media and sends the decoded
// data to che channels to be played.
func (vp *VideoPlay) ReadVideoAndAudio(media *Media) (<-chan *image.RGBA, <-chan [2]float64, chan error, error) {
	frameBuffer := make(chan *image.RGBA, vp.FrameBufferSize)
	sampleBuffer := make(chan [2]float64, vp.SampleBufferSize)
	errs := make(chan error)

	err := media.OpenDecode()

	if err != nil {
		return nil, nil, nil, err
	}

	var videoStream *VideoStream
	if videoStreams := media.VideoStreams(); len(videoStreams) > 0 {
		videoStream = media.VideoStreams()[0]
		err = videoStream.Open()
		if err != nil {
			return nil, nil, nil, err
		}
	}

	var audioStream *AudioStream
	if audioStreams := media.AudioStreams(); len(audioStreams) > 0 {
		audioStream = audioStreams[0]
		err = audioStream.Open()
		if err != nil {
			return nil, nil, nil, err
		}
	}

	/*err = media.Streams()[0].Rewind(60 * time.Second)

	if err != nil {
		return nil, nil, nil, err
	}*/

	/*err = media.Streams()[0].ApplyFilter("h264_mp4toannexb")

	if err != nil {
		return nil, nil, nil, err
	}*/

	go func() {
		for {
			packet, gotPacket, err := media.ReadPacket()

			if err != nil {
				go func(err error) {
					errs <- err
				}(err)
			}

			if !gotPacket {
				break
			}

			/*hash := sha256.Sum256(packet.Data())
			fmt.Println(base58.Encode(hash[:]))*/

			switch packet.Type() {
			case StreamVideo:
				s := media.Streams()[packet.StreamIndex()].(*VideoStream)
				videoFrame, gotFrame, err := s.ReadVideoFrame()

				if err != nil {
					go func(err error) {
						errs <- err
					}(err)
				}

				if !gotFrame {
					break
				}

				if videoFrame == nil {
					continue
				}

				frameBuffer <- videoFrame.Image()

			case StreamAudio:
				if !vp.AudioOn {
					continue
				}
				s := media.Streams()[packet.StreamIndex()].(*AudioStream)
				audioFrame, gotFrame, err := s.ReadAudioFrame()

				if err != nil {
					go func(err error) {
						errs <- err
					}(err)
				}

				if !gotFrame {
					break
				}

				if audioFrame == nil {
					continue
				}

				// Turn the raw byte data into
				// audio samples of type [2]float64.
				reader := bytes.NewReader(audioFrame.Data())

				for reader.Len() > 0 {
					sample := [2]float64{0, 0}
					var result float64
					err = binary.Read(reader, binary.LittleEndian, &result)

					if err != nil {
						go func(err error) {
							errs <- err
						}(err)
					}

					sample[0] = result

					err = binary.Read(reader, binary.LittleEndian, &result)

					if err != nil {
						go func(err error) {
							errs <- err
						}(err)
					}

					sample[1] = result
					sampleBuffer <- sample
				}
			}
		}

		if videoStream != nil {
			videoStream.Close()
		}
		if audioStream != nil {
			audioStream.Close()
		}
		media.CloseDecode()
		close(frameBuffer)
		close(sampleBuffer)
		close(errs)
	}()

	return frameBuffer, sampleBuffer, errs, nil
}

// streamSamples creates a new custom streamer for
// playing audio samples provided by the source channel.
//
// See https://github.com/faiface/beep/wiki/Making-own-streamers
// for reference.
func streamSamples(sampleSource <-chan [2]float64) beep.Streamer {
	return beep.StreamerFunc(func(samples [][2]float64) (n int, ok bool) {
		numRead := 0

		for i := 0; i < len(samples); i++ {
			sample, ok := <-sampleSource

			if !ok {
				numRead = i + 1
				break
			}

			samples[i] = sample
			numRead++
		}

		if numRead < len(samples) {
			return numRead, false
		}

		return numRead, true
	})
}

// converts []uint8 to []uin32
func aUint8to32(pixels []uint8) (data []uint32) {
	// should be faster than using buffer encoding
	if n := len(pixels) / 4; n != 0 {
		data = make([]uint32, n)
		order := binary.LittleEndian
		for i := range data {
			data[i] = order.Uint32(pixels[4*i:])
		}
	}
	return data
}
