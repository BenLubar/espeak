#!/bin/bash -ex

cd build
[[ -d emsdk/.git ]] || git clone --depth=1 https://github.com/juj/emsdk.git
[[ -d espeak-ng/.git ]] || git clone --depth=1 -b 1.49.2 https://github.com/espeak-ng/espeak-ng.git

cd emsdk
git pull
./emsdk install latest
./emsdk activate --embedded latest

cd ../espeak-ng
git fetch origin 1.49.2
git checkout -f 1.49.2
git reset --hard
git clean -fdx

./autogen.sh
./configure --prefix=/usr --without-async --without-mbrola --without-sonic
make

source ../emsdk/emsdk_env.sh

make clean
emconfigure ./configure --prefix=/usr --without-async --without-mbrola --without-sonic
emmake make src/libespeak-ng.la

emcc -o ../../libespeak-ng.inc.js \
	src/.libs/libespeak-ng.a \
	../get_stderr.c \
	-s MODULARIZE=1 \
	-s EXPORTED_FUNCTIONS='[
		"_get_stderr",
		"_espeak_SetSynthCallback",
		"_espeak_ng_ClearErrorContext",
		"_espeak_ng_GetSampleRate",
		"_espeak_ng_GetStatusCodeMessage",
		"_espeak_ng_Initialize",
		"_espeak_ng_InitializeOutput",
		"_espeak_ng_InitializePath",
		"_espeak_ng_PrintStatusCodeMessage",
		"_espeak_ng_SetParameter",
		"_espeak_ng_SetVoiceByName",
		"_espeak_ng_SetVoiceByProperties",
		"_espeak_ng_Synthesize"
	]' \
	-s RESERVED_FUNCTION_POINTERS=1 \
	-s EXTRA_EXPORTED_RUNTIME_METHODS='["addFunction","getValue","setValue"]' \
	-s LZ4=1 \
	-s EXPORT_NAME='"ESpeakNG"' \
	--embed-file espeak-ng-data@/usr/share/espeak-ng-data \
	--exclude-file espeak-ng-data/mbrola_ph \
	--exclude-file espeak-ng-data/phondata-manifest \
	--memory-init-file 0 \
	-Oz

echo 'this.ESpeakNG = ESpeakNG;' >> ../../libespeak-ng.inc.js

cd ../..
echo '// +build js' > js_defs.gen.go
GOARCH=386 go tool cgo -godefs js_defs.go >> js_defs.gen.go
rm -rf _obj
