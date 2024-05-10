working=$(pwd)
mkdir -p $working/source/build/ffmpeg/linux/
cd $working/source/build/ffmpeg/linux/
mkdir $working/lib/linux/
CC=gcc
$working/source/FFmpeg/configure \
	--prefix=$working/lib/linux/ \
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