package webp

/*
#include <stdlib.h>
#include <webp/mux.h>
#include <webp/encode.h>
*/
import "C"
import (
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"math"
	"os"
	"runtime"
	"unsafe"

	_ "golang.org/x/image/tiff"
	_ "golang.org/x/image/webp"
)

type WebPConfig struct {
	enc      *C.WebPAnimEncoder
	config   C.WebPConfig
	wrt      *C.WebPMemoryWriter
	pic      C.WebPPicture
	animated bool
	pinner   runtime.Pinner
}

func (c *WebPConfig) init(config *Config) (err error) {
	if C.WebPConfigInit(&c.config) != 1 {
		err = fmt.Errorf("WebPConfigInit failed")
		return
	}
	c.config.quality = C.float(config.Quality)
	c.config.lossless = C.int(config.Lossless)
	c.config.method = C.int(config.Speed)
	c.config.pass = C.int(config.Pass)
	if config.Targetsize {
		if config.Size > 0 {
			c.config.target_size = C.int(config.Size)
			c.config.quality = C.float(100)
		} else {
			c.config.target_size = C.int(config.Width * config.Height / 10)
			c.config.quality = C.float(100)
		}
	}
	if !c.animated {
		if C.WebPPictureInit(&c.pic) != 1 {
			err = fmt.Errorf("WebPPictureInit failed")
			return
		}
		c.pic.width = C.int(config.Width)
		c.pic.height = C.int(config.Height)
		if C.WebPPictureAlloc(&c.pic) != 1 {
			err = fmt.Errorf("WebPPictureAlloc failed")
			return
		}
		c.pic.use_argb = 1
		c.pic.argb_stride = c.pic.width
		c.wrt = new(C.WebPMemoryWriter)
		c.pinner.Pin(c.wrt)
		C.WebPMemoryWriterInit(c.wrt)
		c.pic.custom_ptr = unsafe.Pointer(c.wrt)
		c.pic.writer = (C.WebPWriterFunction)(C.WebPMemoryWrite)
	} else {
		c.enc = C.WebPAnimEncoderNew(C.int(config.Width), C.int(config.Height), nil)
		if c.enc == nil {
			err = fmt.Errorf("error initializing encoder")
			return
		}
	}
	return
}
func (c *WebPConfig) free() {
	if c.animated {
		C.WebPAnimEncoderDelete(c.enc)
	} else {
		C.WebPPictureFree(&c.pic)
		C.WebPMemoryWriterClear(c.wrt)
		c.pinner.Unpin()
	}
}
func (c *WebPConfig) animAddFrame(config Config, y, u, v *C.uint8_t) (err error) {
	if y == nil {
		ok := C.WebPAnimEncoderAdd(c.enc, nil, C.int(config.timestamp), &c.config)
		if ok != 1 {
			err = fmt.Errorf("error adding frame")
		}
		return
	}
	var pic C.WebPPicture
	if C.WebPPictureInit(&pic) != 1 {
		err = fmt.Errorf("WebPPictureInit failed")
		return
	}
	pic.width = C.int(config.Width)
	pic.height = C.int(config.Height)
	pic.use_argb = 0
	if C.WebPPictureAlloc(&pic) != 1 {
		err = fmt.Errorf("WebPPictureAlloc failed")
		return
	}
	defer C.WebPPictureFree(&pic)
	pic.colorspace = C.WEBP_YUV420
	pic.y, pic.u, pic.v = y, u, v
	pic.y_stride = C.int(config.Width)
	pic.uv_stride = C.int(config.Width / 2)
	if C.WebPAnimEncoderAdd(c.enc, &pic, C.int(config.timestamp), &c.config) != 1 {
		err = fmt.Errorf("error adding frame")
	}
	return
}
func (c *WebPConfig) animExtract() (data []byte, err error) {
	var output C.WebPData
	if C.WebPAnimEncoderAssemble(c.enc, &output) != 1 {
		err = fmt.Errorf("error assembling output")
	}
	data = make([]byte, output.size)
	tempdata := unsafe.Slice((output.bytes), output.size)
	for i, v := range tempdata {
		data[i] = byte(v)
	}
	return
}
func EncodeImage(config Config) (output []byte, err error) {
	file, err := os.Open(config.Filepath)
	if err != nil {
		return
	}
	defer file.Close()
	img, _, err := image.Decode(file)
	if err != nil {
		return
	}
	bounds := img.Bounds()
	config.Width, config.Height = bounds.Max.X, bounds.Max.Y
	data := make([]C.uint32_t, config.Width*config.Height)
	i := 0
	for y := 0; y < config.Height; y++ {
		for x := 0; x < config.Width; x++ {
			r, g, b, a := img.At(x, y).RGBA()
			aF := uint32(math.Round(float64(a) / 257))
			rF := uint32(math.Round(float64(r) / 257))
			gF := uint32(math.Round(float64(g) / 257))
			bF := uint32(math.Round(float64(b) / 257))
			value := aF<<24 | rF<<16 | gF<<8 | bF
			data[i] = C.uint32_t(value)
			i++
		}
	}
	var c WebPConfig
	err = c.init(&config)
	if err != nil {
		return
	}
	defer c.free()
	dataP := unsafe.SliceData(data)
	c.pinner.Pin(dataP)
	c.pic.argb = dataP
	if C.WebPEncode(&c.config, &c.pic) != 1 {
		err = fmt.Errorf("WebPEncode failed:%v", c.pic.error_code)
		return
	}
	output = make([]byte, c.wrt.size)
	Cdata := unsafe.Slice(c.wrt.mem, c.wrt.size)
	for i, v := range Cdata {
		output[i] = byte(v)
	}
	return
}
