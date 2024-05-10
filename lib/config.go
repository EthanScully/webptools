package webp

/*
#cgo linux CFLAGS: -Ilinux/include
#cgo windows CFLAGS: -Iwin/include
#cgo LDFLAGS: -static
#cgo LDFLAGS: -lwebpmux -lwebp -lsharpyuv
#cgo LDFLAGS: -lavformat -lswresample -lswscale -lavcodec -lavutil
#cgo linux LDFLAGS: -Llinux/lib -lz -lm
#cgo windows LDFLAGS: -Lwin/lib -lbcrypt
*/
import "C"
import (
	"os"
	"strings"
)

type Config struct {
	Quality    float64
	Lossless   int
	Speed      int
	Frametime  float64 // in ms
	Width      int
	Height     int
	timestamp  int
	Filepath   string
	Pass       int
	Targetsize bool
	Size       int
}

func (c Config) Name() (filename string) {
	a := strings.Split(c.Filepath, ".")
	b := a[:len(a)-1]
	path := strings.Join(b, ".")
	d := strings.Split(path, string(os.PathSeparator))
	filename = d[len(d)-1] + ".webp"
	return
}
