working=$(pwd)
mkdir -p $working/source/build/webp/win/
cd $working/source/build/webp/win/
mkdir $working/lib/linux/
CC=x86_64-w64-mingw32-gcc
$working/source/libwebp/configure \
	--prefix=$working/lib/win/ \
	--exec-prefix=$working/lib/win/ \
	--host=x86_64-w64-mingw32 \
	--build=x86_64-pc-linux-gnu \
	--target=x86_64-w64-mingw32 \
	--enable-static=yes \
	--enable-shared=no \
	--disable-libwebpdemux
make install