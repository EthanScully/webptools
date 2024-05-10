package webp

/*
#include <libavcodec/avcodec.h>
#include <libavformat/avformat.h>
#include <libavutil/imgutils.h>
#include <libswscale/swscale.h>
*/
import "C"
import (
	"fmt"
	"math"
	"unsafe"
)

type FFmpegConfig struct {
	fmt     *C.AVFormatContext
	codec   *C.AVCodecContext
	pkt     *C.AVPacket
	frame   *C.AVFrame
	index   C.int
	ftconst float64
}

func (d *FFmpegConfig) free() {
	C.av_frame_free(&d.frame)
	C.av_packet_free(&d.pkt)
	C.avcodec_free_context(&d.codec)
	C.avformat_close_input(&d.fmt)
}
func (ff *FFmpegConfig) init(config *Config) (err error) {
	filename := "file:" + config.Filepath
	ff.fmt = nil
	if C.avformat_open_input(&ff.fmt, C.CString(filename), nil, nil) < 0 {
		err = fmt.Errorf("error opening input")
		return
	}
	if C.avformat_find_stream_info(ff.fmt, nil) < 0 {
		err = fmt.Errorf("error finding stream info")
		return
	}
	ff.codec = nil
	ff.index = C.av_find_best_stream(ff.fmt, 0, -1, -1, nil, 0)
	if ff.index < 0 {
		err = fmt.Errorf("could not find stream in input file")
		return
	}
	stream := unsafe.Slice(ff.fmt.streams, ff.index+1)[ff.index]
	ff.ftconst = float64(stream.time_base.num) / float64(stream.time_base.den) * 1000
	dec := C.avcodec_find_decoder(stream.codecpar.codec_id)
	if dec == nil {
		err = fmt.Errorf("error finding codec")
		return
	}
	ff.codec = C.avcodec_alloc_context3(dec)
	if ff.codec == nil {
		err = fmt.Errorf("error allocating codec context")
		return
	}
	if C.avcodec_parameters_to_context(ff.codec, stream.codecpar) < 0 {
		err = fmt.Errorf("error copying codec parameters to context")
		return
	}
	if C.avcodec_open2(ff.codec, dec, nil) < 0 {
		err = fmt.Errorf("error opening codec")
		return
	}
	config.Width = int(ff.codec.width)
	config.Height = int(ff.codec.height)
	if stream == nil {
		err = fmt.Errorf("could not find stream in the input, aborting")
		return
	}
	ff.frame = C.av_frame_alloc()
	if ff.frame == nil {
		err = fmt.Errorf("error allocating frame")
		return
	}
	ff.pkt = C.av_packet_alloc()
	if ff.pkt == nil {
		err = fmt.Errorf("error allocating packet")
		return
	}
	return
}
func convertFrame(src, dst *C.AVFrame) (err error) {
	swsCtx := C.sws_getContext(src.width, src.height, int32(src.format), dst.width, dst.height, int32(dst.format), C.SWS_LANCZOS, nil, nil, nil)
	if swsCtx == nil {
		err = fmt.Errorf("error getting context")
		return
	}
	ret := C.sws_scale(swsCtx, (**C.uint8_t)(unsafe.Pointer(&src.data)), (*C.int)(unsafe.Pointer(&src.linesize)), 0, src.height, (**C.uint8_t)(unsafe.Pointer(&dst.data)), (*C.int)(unsafe.Pointer(&dst.linesize)))
	if ret < 0 {
		err = fmt.Errorf("error scaling frame")
		return
	}
	C.sws_freeContext(swsCtx)
	return
}
func Animated(config Config, verbose bool) (output []byte, err error) {
	inputW, inputH := config.Width, config.Height
	var decode FFmpegConfig
	err = decode.init(&config)
	if err != nil {
		return
	}
	defer decode.free()
	// Scaling
	if inputW != 0 || inputH != 0 {
		if inputW == 0 {
			scale := float64(inputH) / float64(config.Height)
			config.Width = int(math.Round(float64(config.Width) * scale))
			config.Height = int(math.Round(float64(config.Height) * scale))
		} else if inputH == 0 {
			scale := float64(inputW) / float64(config.Width)
			config.Height = int(math.Round(float64(config.Height) * scale))
			config.Width = int(math.Round(float64(config.Width) * scale))
		} else {
			config.Height = inputH
			config.Width = inputW
		}
	}
	if config.Height%2 == 1 {
		config.Height--
	}
	if config.Width%2 == 1 {
		config.Width--
	}
	// //
	var webpconfig WebPConfig
	webpconfig.animated = true
	err = webpconfig.init(&config)
	if err != nil {
		return
	}
	defer webpconfig.free()
	// Init Frame used for Conversion
	convFrame := C.av_frame_alloc()
	if convFrame == nil {
		err = fmt.Errorf("error allocating convert frame")
		return
	}
	defer C.av_frame_free(&convFrame)
	convFrame.height = C.int(config.Height)
	convFrame.width = C.int(config.Width)
	convFrame.format = C.AV_PIX_FMT_YUV420P
	ret := C.av_image_alloc((**C.uint8_t)(unsafe.Pointer(&convFrame.data)), (*C.int)(unsafe.Pointer(&convFrame.linesize)), convFrame.width, convFrame.height, int32(convFrame.format), 1)
	if ret < 0 {
		err = fmt.Errorf("error allocating frame data for conversion frame; av_image_alloc code:%v", ret)
		return
	}
	defer C.av_freep(unsafe.Pointer(&convFrame.data[0]))
	// Verbose
	var frames int
	frame := 0
	if verbose {
		frames, err = GetTotalFrames(config)
		if err != nil {
			return
		}
	}
	// //
	var decodeEncode func(pkt *C.AVPacket)
	for C.av_read_frame(decode.fmt, decode.pkt) >= 0 {
		if decode.pkt.stream_index == decode.index {
			decodeEncode = func(pkt *C.AVPacket) {
				if C.avcodec_send_packet(decode.codec, pkt) < 0 {
					err = fmt.Errorf("error sending packet")
					return
				}
				for ret >= 0 {
					ret = C.avcodec_receive_frame(decode.codec, decode.frame)
					if ret < 0 {
						if ret == C.AVERROR_EOF || ret == -11 {
							ret = 0
							return
						}
						err = fmt.Errorf("error durring decoding")
						return
					}
					err = convertFrame(decode.frame, convFrame)
					if err != nil {
						return
					}
					y := convFrame.data[0]
					u := convFrame.data[1]
					v := convFrame.data[2]
					err = webpconfig.animAddFrame(config, y, u, v)
					if err != nil {
						return
					}
					if config.Frametime == 0 {
						config.timestamp += int(float64(decode.frame.duration) * decode.ftconst)
					} else {
						config.timestamp += int(config.Frametime)
					}
					C.av_frame_unref(decode.frame)
					// verbose
					if verbose {
						frame++
						fmt.Printf("\n\033[A%.1f%%", 100*float64(frame)/float64(frames))
					}
					// //
				}
			}
			decodeEncode(decode.pkt)
			if err != nil {
				return
			}
		}
		C.av_packet_unref(decode.pkt)
	}
	// flush decoder
	decodeEncode(nil)
	if err != nil {
		return
	}
	// flush encoder
	err = webpconfig.animAddFrame(config, nil, nil, nil)
	if err != nil {
		return
	}
	output, err = webpconfig.animExtract()
	if err != nil {
		return
	}
	if verbose {
		fmt.Println("")
	}
	return
}
func GetTotalFrames(config Config) (frames int, err error) {
	var decode FFmpegConfig
	err = decode.init(&config)
	if err != nil {
		return
	}
	defer decode.free()
	for C.av_read_frame(decode.fmt, decode.pkt) >= 0 {
		if decode.pkt.stream_index == decode.index {
			frames++
		}
		C.av_packet_unref(decode.pkt)
	}
	return
}
