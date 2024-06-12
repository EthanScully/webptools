package webp

/*
#cgo linux,amd64 CFLAGS: -Ilinux-amd64/include
#cgo windows,amd64 CFLAGS: -Iwindows-amd64/include
#cgo linux,arm64 CFLAGS: -Ilinux-arm64/include
#cgo windows,arm64 CFLAGS: -Iwindows-arm64/include
#cgo LDFLAGS: -static
#cgo LDFLAGS: -lwebpmux -lwebp -lsharpyuv
#cgo LDFLAGS: -lavformat -lswresample -lswscale -lavcodec -lavutil
#cgo linux,amd64 LDFLAGS: -Llinux-amd64/lib -lz -lm
#cgo windows,amd64 LDFLAGS: -Lwindows-amd64/lib -lbcrypt
#cgo linux,arm64 LDFLAGS: -Llinux-arm64/lib -lm
#cgo windows,arm64 LDFLAGS: -Lwindows-arm64/lib -lbcrypt
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
