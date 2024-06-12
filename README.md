## Create Animated WEBPs from Video using libwebp and FFmpeg in Go

### For Usage Details
```POWERSHELL
.\webp-tools-windows-amd64.exe --help
```
This program utilizes the [FFmpeg](https://github.com/FFmpeg/FFmpeg) and [libwebp](https://github.com/webmproject/libwebp) C api with CGO to create Webp files
## Build
only supports building on linux right now

when in repository directory:
```BASH
go run build/*
go run build/* --arch arm64
go run build/* --os windows
go run build/* --os windows --arch arm64
```
### Docker
```BASH
docker run --rm \
    -v $(pwd):/mnt/ -w /mnt/ \
    -t debian:sid bash -c '
    apt update
    apt install build-essential make nasm yasm zlib1g-dev liblzma-dev golang ca-certificates tar gcc-aarch64-linux-gnu gcc autoconf automake libtool -y
    go run build/* mingw
    go run build/*
    go run build/* --arch arm64
    go run build/* --os windows
    go run build/* --os windows --arch arm64'
```