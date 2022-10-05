SDL2 FFmpeg 
===========
A simple go library for playing video and audio using the ffmpeg for decoding the video/audio and SDL2 for rendering the video.
The library is based on [Reisen](https://github.com/zergon321/reisen) library which is based on **libav** (i.e. **ffmpeg**).

[![Build Status](https://github.com/boseca/go-sdl2-ffmpeg/workflows/build/badge.svg)](https://github.com/boseca/go-sdl2-ffmpeg/actions?query=workflow%3Abuild)
[![Coverage Status](https://coveralls.io/repos/github/boseca/go-sdl2-ffmpeg/badge.svg?branch=master)](https://coveralls.io/github/boseca/go-sdl2-ffmpeg?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/boseca/go-sdl2-ffmpeg?1)](https://goreportcard.com/report/github.com/boseca/go-sdl2-ffmpeg)
[![Go Reference](https://pkg.go.dev/badge/github.com/boseca/go-sdl2-ffmpeg.svg)](https://pkg.go.dev/github.com/boseca/go-sdl2-ffmpeg)

## Dependencies
Following components are required for this library to work.

>Please note MSYS2 is required only if you need to build the project for Windows

- [Go v1.17+](https://go.dev/dl/)
    - Windows   
        GO should be installed on Windows OS when MSYS2 is used

        - download and install on Windows OS (see [Go install](https://go.dev/doc/install))

        - set GO path in MSYS2   
            `echo 'export PATH=/c/Program\ Files/Go/bin:$PATH' >> ~/.bashrc`
        
        - build GO package example
            ```bash
            gi clone https://github.com/faiface/pixel-examples.git
            cd platformer
            go run main.go
            go build
            ./platformer.exe
            ```

    - Linux
        GO should be installed in default folder (see [Go install](https://go.dev/doc/install))
        ```bash
        rm -rf /usr/local/go && tar -C /usr/local -xzf go1.19.1.linux-amd64.tar.gz
        export PATH=$PATH:/usr/local/go/bin
        go version
        ```

- [MSYS2](http://www.msys2.org/) (for Windows)

- [GCC toolchain](https://gcc.gnu.org/)
    - Linux
        ```bash
        sudo apt install build-essential libtool autotools-dev automake pkg-config bsdmainutils curl git
        gcc --version
        ```
    - Windows (MSYS2)
        ```bash
        pacman -S --needed base-devel mingw-w64-i686-toolchain mingw-w64-x86_64-toolchain        
        echo 'export PATH=/mingw64/bin:$PATH' >> ~/.bashrc
        pacman -S git
        echo 'export PATH=/c/Go/bin:$PATH' >> ~/.bashrc # set path to a default Go installation
        # OPTIONAL
        echo 'export GOPATH=$USERPROFILE/go' >> ~/.bashrc
        ```

- [SDL2](http://libsdl.org/download-2.0.php)
    - [go-sdl2](https://github.com/veandco/go-sdl2)    
        - [requirements](https://github.com/veandco/go-sdl2#requirements)  
            SDL2 libraries has to be installed for `go-sdl2` to work

            - For Ubuntu 22.04 and above  
                `apt install libsdl2{,-image,-mixer,-ttf,-gfx}-dev`

            - For Windows (MSYS2)   
                `pacman -S mingw-w64-x86_64-gcc mingw-w64-x86_64-SDL2{,_image,_mixer,_ttf,_gfx}`
        
        - [Installation](https://github.com/veandco/go-sdl2#installation)  
            ```bash
            go get -v github.com/veandco/go-sdl2/sdl
            go get -v github.com/veandco/go-sdl2/img # optional
            go get -v github.com/veandco/go-sdl2/mix # optional
            go get -v github.com/veandco/go-sdl2/ttf # optional
            go get -v github.com/veandco/go-sdl2/gfx # optional
            ```
- [FFmpeg]
    - Linux
        - [Libav](https://libav.org/) (ffmpeg)
            - **libavformat**
            - **libavcodec**
            - **libavutil**
            - **libswresample**
            - **libswscale**
            - **libasound2**

        For **Arch**-based **Linux** distributions:

        ```bash
        sudo pacman -S ffmpeg
        ```

        For **Debian**-based **Linux** distributions:

        ```bash
        sudo apt install libswscale-dev libavcodec-dev libavformat-dev libswresample-dev libavutil-dev libasound2-dev
        sudo apt install libgl1-mesa-dev xorg-dev # adds X11/Xlib.h
        ```

        For **macOS**:

        ```bash
        brew install libav
        ```

    - Windows (MSYS2)
        `pacman -S mingw-w64-x86_64-ffmpeg`

    For **Windows** see the [detailed tutorial](https://medium.com/@maximgradan/how-to-easily-bundle-your-cgo-application-for-windows-8515d2b19f1e).

## Installation

```bash
go get github.com/boseca/go-sdl2-ffmpeg
```

## Packaging 
Copy all binaaries necessary to run the applicaiton into the `bundle` folder
`ldd player.exe | python bundle.py`

## Test

### Run
Run tests for `sfplay` package
`go test -v github.com/boseca/go-sdl2-ffmpeg/sfplay`

### Build
Build the test project and play a demo video
- Linux
    `cd examples/player && go build -x -o player *.go && ./run.sh`
- Windows (MSYS2)
    `cd examples/player && go build -ldflags "-s -w -H=windowsgui" && ./run.bat`
>Info
    -s and -w options are required to strip the unneeded debug symbols from the executable and 
    -H is required to suppress the launch of the command line when we start the application.

## Usage

Any media file is composed of streams containing media data, e.g. audio, video and subtitles. The whole presentation data of the file is divided into packets. Each packet belongs to one of the streams and represents a single frame of its data. The process of decoding implies reading packets and decoding them into either video frames or audio frames.

The library provides read video frames as **RGBA** pictures. The audio samples are provided as raw byte slices in the format of `AV_SAMPLE_FMT_DBL` (i.e. 8 bytes per sample for one channel, the data type is `float64`). The channel layout is stereo (2 channels). The byte order is little-endian. The detailed scheme of the audio samples sequence is given below.

![Audio sample structure](./pictures/audio_sample_structure.png)

Once the frame is decoded the SDL texture is updated with the new frame pixels and SDL event is triggered to render the frame in the SDL window.

You are welcome to look at the [examples](./examples) to understand how to work with the library. Also please take a look at the detailed [tutorial](https://medium.com/@maximgradan/playing-videos-with-golang-83e67447b111).
