working=$(pwd)
mkdir -p $working/source/build/webp/linux/
cd $working/source/build/webp/linux/
mkdir $working/lib/linux/
CC=gcc
$working/source/libwebp/configure \
	--prefix=$working/lib/linux/ \
	--exec-prefix=$working/lib/linux/ \
	--enable-static=yes \
	--enable-shared=no \
	--disable-libwebpdemux
make install