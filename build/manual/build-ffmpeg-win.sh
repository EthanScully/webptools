working=$(pwd)
mkdir -p $working/source/build/ffmpeg/win/
cd $working/source/build/ffmpeg/win/
mkdir $working/lib/win/
CC=x86_64-w64-mingw32-gcc
$working/source/FFmpeg/configure \
	--target-os=mingw32 \
	--cross-prefix=x86_64-w64-mingw32- \
	--arch=x86_64 \
	--enable-cross-compile \
	--prefix=$working/lib/win/ \
	--disable-avdevice \
	--disable-postproc \
	--disable-avfilter \
	--disable-doc \
	--disable-htmlpages \
	--disable-manpages \
	--disable-podpages \
	--disable-txtpages \
	--disable-programs \
	--disable-network \
	--disable-everything \
	--enable-decoder=av1 \
	--enable-decoder=hevc \
	--enable-decoder=h264 \
	--enable-decoder=rawvideo \
	--enable-decoder=dnxhd \
	--enable-decoder=prores \
	--enable-demuxer=mov \
	--enable-demuxer=matroska \
	--enable-demuxer=m4v \
	--enable-parser=av1 \
	--enable-parser=hevc \
	--enable-parser=h264 \
	--enable-parser=dnxhd \
	--enable-protocol=file
#	--disable-debug 
make install