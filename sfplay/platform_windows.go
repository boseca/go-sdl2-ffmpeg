package sfplay

import "C"

func bufferSize(maxBufferSize C.int) C.ulonglong {
	var byteSize C.ulonglong = 8
	return C.ulonglong(maxBufferSize) * byteSize
}

func channelLayout(audio *AudioStream) C.longlong {
	return C.longlong(audio.codecCtx.channel_layout)
}

func rewindPosition(dur int64) C.longlong {
	return C.longlong(dur)
}
