package main

// build:
// windows:		CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc GOOS=windows go build -ldflags="-s -w" -o /mnt/c/Users/scull/Desktop/webp.exe webp/cli
// linux:		go build -ldflags="-s -w"

import (
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	webp "webp/lib"

	_ "golang.org/x/image/tiff"
	_ "golang.org/x/image/webp"
)

func BatchConvert(config webp.Config) (err error) {
	if config.Filepath[len(config.Filepath)-2] != string(os.PathSeparator)[0] {
		err = fmt.Errorf("batch conversion requires a directory")
		return
	}
	dir := config.Filepath[:len(config.Filepath)-2]
	files, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	var wg sync.WaitGroup
	var mtx sync.Mutex
	threads := 0
	done := 0
	cores := runtime.NumCPU()
	fmt.Printf("\n\033[A\033[2KProgress: %.1f%%", 0.0)
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		wg.Add(1)
		go func() {
			tempconfig := config
			tempconfig.Filepath = dir + string(os.PathSeparator) + file.Name()
			data, err := webp.EncodeImage(tempconfig)
			if err != nil {
				fmt.Printf("file:%v-error:%v\n", file.Name(), err)
			}
			mtx.Lock()
			done++
			threads--
			mtx.Unlock()
			percent := float64(done) / float64(len(files)) * 100
			fmt.Printf("\n\033[A\033[2KProgress: %.1f%%", percent)
			os.WriteFile(tempconfig.Name(), data, 0666)
			wg.Done()
		}()
		mtx.Lock()
		threads++
		mtx.Unlock()
		for threads >= cores {
			time.Sleep(time.Microsecond * 250)
		}
	}
	wg.Wait()
	return
}
func cliInit() (config webp.Config) {
	arg := func(i int) (arg string, err error) {
		if i < len(os.Args) {
			arg = os.Args[i]
		} else {
			err = fmt.Errorf("missing argument")
		}
		return
	}
	help := func() {
		execName := "<exec>"
		if len(os.Args) > 0 {
			execName = os.Args[0]
			execIndex := strings.LastIndex(execName, string(os.PathSeparator))
			if execIndex != -1 && len(execName) > execIndex+1 {
				execName = execName[execIndex+1:]
			}
		}
		fmt.Printf("Usage: %s <input file> <quality> <speed> <lossless> <targetsize> <pass> <bytes> <framerate> <width> <height>\n", execName)
		fmt.Println("\tinput file:\tinput file to convert")
		fmt.Println("\tquality:\tquality of the output (0-100)")
		fmt.Println("\tspeed:\t\tspeed of the output (0-6)")
		fmt.Println("\tlossless:\tlossless mode (0-1)")
		fmt.Println("\ttargetsize:\ttargetsize mode (0-1)")
		fmt.Println("\tpass:\t\t num of passes for targetsize mode (1-10)")
		fmt.Println("\tbytes:\t\t size per frame in bytes for targetsize mode (>=0), 0=preset")
		fmt.Println("\t\t\tAnimation Only:")
		fmt.Println("\tframerate:\tframerate of the output (1-90), 0=auto")
		fmt.Println("\twidth:\t\tscale width of the output (>=0), 0=auto")
		fmt.Println("\theight:\t\tscale height of the output (>=0), 0=auto")
		os.Exit(0)
	}
	var err error
	i := 1
	config.Filepath, err = arg(i)
	if err != nil {
		help()
	}
	if config.Filepath == "--help" {
		help()
	}
	var temp any
	i++
	framenum, err := arg(i)
	if err == nil {
		if framenum == "frames" {
			totalFrames, err := webp.GetTotalFrames(config)
			if err != nil {
				fmt.Println(err)
				os.Exit(0)
			}
			fmt.Println(totalFrames)
			os.Exit(0)
		}
	}
	// quality
	quality, err := arg(i)
	if err != nil {
		quality = "85"
	}
	config.Quality, err = strconv.ParseFloat(quality, 64)
	if err != nil {
		config.Quality = 85.0
	} else if config.Quality < 0.0 || config.Quality > 100.0 {
		config.Quality = 85.0
	}
	i++
	// speed
	speed, err := arg(i)
	if err != nil {
		speed = "0"
	}
	config.Speed, err = strconv.Atoi(speed)
	if err != nil {
		config.Speed = 0
	} else if config.Speed < 0 || config.Speed > 6 {
		config.Speed = 0
	}
	i++
	// lossless
	lossless, err := arg(i)
	if err != nil {
		config.Lossless = 0
	} else if lossless == "1" {
		config.Lossless = 1
	} else {
		config.Lossless = 0
	}
	i++
	// targetsize
	targetsize, err := arg(i)
	if err != nil {
		config.Targetsize = false
	} else if targetsize == "1" {
		config.Targetsize = true
	} else {
		config.Targetsize = false
	}
	i++
	// pass
	pass, err := arg(i)
	if err != nil {
		if config.Targetsize {
			pass = "6"
		} else {
			pass = "1"
		}
	}
	config.Pass, err = strconv.Atoi(pass)
	if err != nil {
		config.Pass = 1
	} else if config.Pass < 1 || config.Pass > 10 {
		config.Pass = 1
	}
	i++
	// size in bytes
	size, err := arg(i)
	if err != nil {
		size = "0"
	}
	config.Size, err = strconv.Atoi(size)
	if err != nil {
		config.Size = 0
	} else if config.Size <= 0 {
		config.Size = 0
	} else {
		config.Targetsize = true
	}
	i++
	// framerate
	framerate, err := arg(i)
	if err != nil {
		framerate = "0"
	}
	temp, err = strconv.ParseFloat(framerate, 64)
	if err != nil {
		config.Frametime = 0.0
	} else if temp != 0.0 {
		config.Frametime = 1.0 / temp.(float64) * 1000
	}
	if config.Frametime < 11.0 || config.Frametime > 1000.0 {
		config.Frametime = 0.0
	}
	i++
	// width and height
	width, err := arg(i)
	if err != nil {
		width = "0"
	}
	height, err := arg(i + 1)
	if err != nil {
		height = "0"
	}
	scaleW, err := strconv.Atoi(width)
	if err != nil {
		scaleW = 0
	}
	scaleH, err := strconv.Atoi(height)
	if err != nil {
		scaleH = 0
	}
	if scaleH > 0 && scaleW > 0 {
		config.Width = scaleW
		config.Height = scaleH
	} else if scaleH > 0 {
		config.Height = scaleH
		config.Width = 0
	} else if scaleW > 0 {
		config.Width = scaleW
		config.Height = 0
	} else {
		config.Width = 0
		config.Height = 0
	}
	i += 2
	return
}
func main() {
	config := cliInit()
	var filetype string
	if config.Filepath[len(config.Filepath)-1] == "*"[0] {
		err := BatchConvert(config)
		if err != nil {
			panic(err)
		}
		return
	}
	file, err := os.Open(config.Filepath)
	if err != nil {
		panic(err)
	}
	_, filetype, _ = image.Decode(file)
	file.Close()
	var data []byte
	if filetype == "png" || filetype == "jpeg" || filetype == "tiff" || filetype == "webp" {
		data, err = webp.EncodeImage(config)
		if err != nil {
			panic(err)
		}
	} else {
		data, err = webp.Animated(config, true)
		if err != nil {
			panic(err)
		}
	}
	os.WriteFile(config.Name(), data, 0666)
}
