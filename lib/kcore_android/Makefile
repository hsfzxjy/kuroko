SHELL := /bin/bash

export GOOS=android
export GOARCH=arm64
export CGO_ENABLED=1

NDK_ARCH=aarch64
NDK_AV=31
NDK_DIR=$(shell bash -c 'dirname `ndk-which readelf`')

export CC="$(NDK_DIR)/$(NDK_ARCH)-linux-android$(NDK_AV)-clang"
export CXX="$(NDK_DIR)/$(NDK_ARCH)-linux-android$(NDK_AV)-clang++"
AR="$(NDK_DIR)/llvm-ar"

DART_SDK_INCLUDE_DIR = $(shell bash -c 'dirname `which flutter`')/cache/dart-sdk/include/

OUT_DIR=native/build

.PHONY: all
all: so ffigen

$(OUT_DIR)/kcorec.o: $(wildcard native/c/*.c)
	mkdir -p $(OUT_DIR)
	$(CC) -I$(DART_SDK_INCLUDE_DIR) -llog -c $^ -o $@

$(OUT_DIR)/libkcorec.a: $(OUT_DIR)/kcorec.o
	$(AR) rcs $@ $^

.PHONY: kcorec
kcorec: $(OUT_DIR)/libkcorec.a

GOSRC=$(shell find native -regextype posix-extended -regex '.*\.(c|h|go|mod)')
SO_OUT_DIR=android/libs/arm64-v8a

.PHONY: so
so: $(SO_OUT_DIR)/libkcore.so

$(SO_OUT_DIR)/libkcore.so: $(GOSRC) $(OUT_DIR)/libkcorec.a
	ls -al $?
	cd native/go && \
	 go build -x -a -buildmode=c-shared -o ./libkcore.so && \
	 mkdir -p ../../$(SO_OUT_DIR) && \
	 mv libkcore.so ../../$(SO_OUT_DIR)
	

DART_FFIGEN = $(wildcard lib/*binding.dart)

.PHONY: ffigen
ffigen: $(DART_FFIGEN)

$(DART_FFIGEN): $(SO_OUT_DIR)/libkcore.so pubspec.yaml
	rm include -rf
	for hfile in $$(find native -name '*.h'); do \
		mkdir -p $$(dirname $${hfile/native/include}) ; \
		cp $$hfile $$(dirname $${hfile/native/include}) ; \
	done

	dart run ffigen