package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

var (
	OS, ARCH, AR, CC, OUTPUT string
	workingDir, _            = os.Getwd()
	ffmpegMakeArgs           = []string{
		"disable-avdevice",
		"disable-postproc",
		"disable-avfilter",
		"disable-doc",
		"disable-htmlpages",
		"disable-manpages",
		"disable-podpages",
		"disable-txtpages",
		"disable-programs",
		"disable-network",
		"disable-everything",
		"enable-decoder=av1",
		"enable-decoder=hevc",
		"enable-decoder=h264",
		"enable-decoder=rawvideo",
		"enable-decoder=dnxhd",
		"enable-decoder=prores",
		"enable-demuxer=mov",
		"enable-demuxer=matroska",
		"enable-demuxer=m4v",
		"enable-parser=av1",
		"enable-parser=hevc",
		"enable-parser=h264",
		"enable-parser=dnxhd",
		"enable-protocol=file",
		"disable-debug",
	}
	webpMakeArgs = []string{
		"enable-static=yes",
		"enable-shared=no",
		"disable-libwebpdemux",
	}
)

func arg(i int) (arg string) {
	if len(os.Args) > i {
		arg = os.Args[i]
	}
	return
}
func help() (out string) {
	out = `Supported Arguments:
	--os		windows|linux
	--arch		arm64
	--debug		enable ffmpeg debug mode
	--output	executable output path/name`
	return
}
func sendCmd(name string, arg ...string) error {
	cmd := exec.Command(name, arg...)
	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()
	err := cmd.Start()
	if err != nil {
		return err
	}
	streamOutput := func(reader io.Reader) {
		buffer := make([]byte, 1024)
		for {
			n, err := reader.Read(buffer)
			if err != nil {
				if err != io.EOF {
					fmt.Println("Error reading output:", err)
				}
				break
			}
			if n > 0 {
				fmt.Print(string(buffer[:n]))
			}
		}
	}
	go streamOutput(stdout)
	go streamOutput(stderr)
	err = cmd.Wait()
	if err != nil {
		return err
	}
	return nil
}
func commandExists(cmd string) bool {
	output, err := exec.Command("which", cmd).CombinedOutput()
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(output)) != ""
}
func preCheck() (err error) {
	if runtime.GOOS != "linux" {
		return fmt.Errorf("unsupported build os, linux only")
	}
	_, err = os.Stat("build/main.go")
	if err != nil {
		return fmt.Errorf("this must be run from the root of the repository")
	}
	// parse arguments
	OS = runtime.GOOS
	ARCH = runtime.GOARCH
	for i, v := range os.Args {
		switch v[:2] {
		case "--":
			switch v[2:] {
			case "os":
				OS = arg(i + 1)
			case "arch":
				ARCH = arg(i + 1)
			case "debug":
				ret := -1
				for i, v := range ffmpegMakeArgs {
					if v == "disable-debug" {
						ret = i
					}
				}
				if ret != -1 {
					ffmpegMakeArgs = append(ffmpegMakeArgs[:ret], ffmpegMakeArgs[ret+1:]...)
				}
			case "help":
				fmt.Println(help())
			case "output":
				OUTPUT = arg(i + 1)
			}
		}
	}
	// Make Setup
	switch OS {
	case "linux":
		ffmpegMakeArgs = append([]string{"target-os=linux"}, ffmpegMakeArgs...)
		runtimeArch := runtime.GOARCH
		switch ARCH {
		case runtimeArch:
			CC = "gcc"
			AR = "ar"
		case "amd64":
			CC = "x86_64-linux-gnu-gcc"
			AR = "x86_64-linux-gnu-ar"
			ffmpegMakeArgs = append([]string{"cross-prefix=x86_64-linux-gnu-"}, ffmpegMakeArgs...)
			webpMakeArgs = append([]string{"host=x86_64-linux-gnu"}, webpMakeArgs...)
			webpMakeArgs = append([]string{"target=x86_64-linux-gnu"}, webpMakeArgs...)
		case "arm64":
			AR = "aarch64-linux-gnu-ar"
			CC = "aarch64-linux-gnu-gcc"
			ffmpegMakeArgs = append([]string{"cross-prefix=aarch64-linux-gnu-"}, ffmpegMakeArgs...)
			webpMakeArgs = append([]string{"host=aarch64-linux-gnu"}, webpMakeArgs...)
			webpMakeArgs = append([]string{"target=aarch64-linux-gnu"}, webpMakeArgs...)
		default:
			return fmt.Errorf("amd64 and arm64 support only")
		}
	case "windows":
		ffmpegMakeArgs = append([]string{"target-os=mingw32"}, ffmpegMakeArgs...)
		switch ARCH {
		case "amd64":
			AR = "x86_64-w64-mingw32-ar"
			CC = "x86_64-w64-mingw32-gcc"
			ffmpegMakeArgs = append([]string{"cross-prefix=x86_64-w64-mingw32-"}, ffmpegMakeArgs...)
			webpMakeArgs = append([]string{"host=x86_64-w64-mingw32"}, webpMakeArgs...)
			webpMakeArgs = append([]string{"target=x86_64-w64-mingw32"}, webpMakeArgs...)
		case "arm64":
			AR = "aarch64-w64-mingw32-ar"
			CC = "aarch64-w64-mingw32-gcc"
			ffmpegMakeArgs = append([]string{"cross-prefix=aarch64-w64-mingw32-"}, ffmpegMakeArgs...)
			webpMakeArgs = append([]string{"host=aarch64-w64-mingw32"}, webpMakeArgs...)
			webpMakeArgs = append([]string{"target=aarch64-w64-mingw32"}, webpMakeArgs...)
		default:
			return fmt.Errorf("amd64 and arm64 support only")
		}
	default:
		return fmt.Errorf("linux and windows support only")
	}
	switch ARCH {
	case "arm64":
		ffmpegMakeArgs = append([]string{"arch=aarch64"}, ffmpegMakeArgs...)
	case "amd64":
		ffmpegMakeArgs = append([]string{"arch=x86_64"}, ffmpegMakeArgs...)
	}
	if OS != runtime.GOOS || ARCH != runtime.GOARCH {
		ffmpegMakeArgs = append([]string{"enable-cross-compile"}, ffmpegMakeArgs...)
		switch runtime.GOARCH {
		case "amd64":
			webpMakeArgs = append([]string{"build=x86_64-linux-gnu"}, webpMakeArgs...)
		case "arm64":
			webpMakeArgs = append([]string{"build=aarch64-linux-gnu"}, webpMakeArgs...)
		}
	}
	// Set Environment Variables
	os.Setenv("GOOS", OS)
	os.Setenv("GOARCH", ARCH)
	os.Setenv("CC", CC)
	os.Setenv("AR", AR)
	os.Setenv("CGO_ENABLED", "1")
	return
}
func buildFFmpeg() (err error) {
	// Soft Dependency Check
	ok := commandExists(CC)
	if !ok {
		return fmt.Errorf("%s not found", CC)
	}
	ok = commandExists("make")
	if !ok {
		return fmt.Errorf("make not found")
	}

	// Create Build Directory
	buildDir := fmt.Sprintf("%s/bin/FFmpeg/%s-%s", workingDir, OS, ARCH)
	err = os.MkdirAll(buildDir, 0770)
	if err != nil && !strings.Contains(err.Error(), "exists") {
		return
	}
	// Create Libriary Directory
	libDir := fmt.Sprintf("%s/lib/%s-%s", workingDir, OS, ARCH)
	// Set install directory of Make
	ffmpegMakeArgs = append([]string{"prefix=" + libDir}, ffmpegMakeArgs...)
	// Build FFmpeg
	err = os.Chdir(buildDir)
	if err != nil {
		return
	}
	defer os.Chdir(workingDir)
	var cmd []string
	for _, v := range ffmpegMakeArgs {
		cmd = append(cmd, "--"+v)
	}
	err = sendCmd(fmt.Sprintf("%s/source/FFmpeg/configure", workingDir), cmd...)
	if err != nil {
		return fmt.Errorf("make configure failed: %s\nnasm/yasm might need to be installed", err)
	}
	err = sendCmd("make", fmt.Sprintf("-j%d", runtime.NumCPU()))
	if err != nil {
		return fmt.Errorf("make failed: %s", err)
	}
	err = os.Mkdir(libDir, 0770)
	if err != nil && !strings.Contains(err.Error(), "exists") {
		return
	}
	err = sendCmd("make", "install")
	if err != nil {
		sendCmd("rm", "-rf", libDir)
		return fmt.Errorf("make install failed: %s", err)
	}
	return
}
func buildWebp() (err error) {
	// Soft Dependency Check
	ok := commandExists(CC)
	if !ok {
		return fmt.Errorf("%s not found", CC)
	}
	ok = commandExists("make")
	if !ok {
		return fmt.Errorf("make not found")
	}

	// Create Build Directory
	buildDir := fmt.Sprintf("%s/bin/webp/%s-%s", workingDir, OS, ARCH)
	err = os.MkdirAll(buildDir, 0770)
	if err != nil && !strings.Contains(err.Error(), "exists") {
		return
	}
	// Create Libriary Directory
	libDir := fmt.Sprintf("%s/lib/%s-%s", workingDir, OS, ARCH)
	// Set install directory of Make
	webpMakeArgs = append([]string{"prefix=" + libDir}, webpMakeArgs...)
	webpMakeArgs = append([]string{"exec-prefix=" + libDir}, webpMakeArgs...)
	// Run autoconf
	err = os.Chdir(fmt.Sprintf("%s/source/libwebp", workingDir))
	if err != nil {
		return
	}
	defer os.Chdir(workingDir)
	err = sendCmd("autoreconf", "-fi")
	if err != nil {
		return fmt.Errorf("autoreconf failed: %s", err)
	}
	// Build libwebp
	err = os.Chdir(buildDir)
	if err != nil {
		return
	}
	var cmd []string
	for _, v := range webpMakeArgs {
		cmd = append(cmd, "--"+v)
	}
	err = sendCmd(fmt.Sprintf("%s/source/libwebp/configure", workingDir), cmd...)
	if err != nil {
		return fmt.Errorf("make configure failed: %s", err)
	}
	err = sendCmd("make", fmt.Sprintf("-j%d", runtime.NumCPU()))
	if err != nil {
		return fmt.Errorf("make failed: %s", err)
	}
	err = os.Mkdir(libDir, 0770)
	if err != nil && !strings.Contains(err.Error(), "exists") {
		return
	}
	err = sendCmd("make", "install")
	if err != nil {
		sendCmd("rm", "-rf", libDir)
		return fmt.Errorf("make install failed: %s", err)
	}
	err = os.Remove(fmt.Sprintf("%s/source/libwebp/configure~", workingDir))
	if err != nil {
		return
	}
	return
}
func main() {
	if arg(1) == "mingw" {
		err := getmingw()
		if err != nil {
			panic(err)
		}
		return
	}
	err := preCheck()
	if err != nil {
		panic(err)
	}
	_, err = os.Stat(fmt.Sprintf("%s/lib/%s-%s", workingDir, OS, ARCH))
	if err != nil {
		err := buildFFmpeg()
		if err != nil {
			panic(err)
		}
		err = buildWebp()
		if err != nil {
			panic(err)
		}
	}
	ouputPath := fmt.Sprintf("bin/webp-tools-%s-%s", OS, ARCH)
	err = os.Mkdir("bin", 0770)
	if err != nil && !strings.Contains(err.Error(), "exists") {
		panic(err)
	}
	if OS == "windows" {
		ouputPath = ouputPath + ".exe"
	}
	if OUTPUT != "" {
		ouputPath = OUTPUT
	}
	err = sendCmd("go", "build", "-ldflags=-s -w", "-o", ouputPath, "cli/main.go")
	if err != nil {
		fmt.Println("look at lib/config.go for any lib requirments")
		panic(err)
	}
}
