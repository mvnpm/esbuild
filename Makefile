## Platform Makefile
##

VERSION=0.23.0-mvnpm-0.0.8
SASS_VERSION=1.78.0

platform-all:
	@$(MAKE) --no-print-directory -j1 \
		platform-darwin-arm64 \
		platform-darwin-x64 \
		platform-linux-arm64 \
		platform-linux-ia32 \
		platform-linux-x64 \
		platform-win32-ia32 \
		platform-win32-x64

NAME = $(GOOS)
ifeq ($(NAME), darwin)
	NAME = macos
endif
download-nix:
	curl -OL https://github.com/sass/dart-sass/releases/download/$(SASS_VERSION)/dart-sass-$(SASS_VERSION)-$(NAME)-$(GOARCH).tar.gz
	tar -xzvf dart-sass-$(SASS_VERSION)-$(NAME)-$(GOARCH).tar.gz
	mv dart-sass $(BUILDDIR)

download-win:
	curl -OL https://github.com/sass/dart-sass/releases/download/$(SASS_VERSION)/dart-sass-$(SASS_VERSION)-$(GOOS)-$(GOARCH).zip
	unzip dart-sass-$(SASS_VERSION)-$(GOOS)-$(GOARCH).zip
	mv dart-sass $(BUILDDIR)

build-win32:
	CGO_ENABLED=0 GOOS="$(GOOS)" GOARCH="$(GOARCH)" go build $(GO_FLAGS) -o "build/$(BUILDDIR)/package/bin/esbuild.exe" ./esbuild_scss.go
	@$(MAKE) --no-print-directory GOOS="$(GOOS)" GOARCH="$(ARCH)" BUILDDIR="build/$(BUILDDIR)" download-win
	tar -czf "build/esbuild-$(BUILDDIR)-$(VERSION).tgz" --directory "build/$(BUILDDIR)" package dart-sass

platform-win32-x64:
	@$(MAKE) --no-print-directory GOOS=windows GOARCH=amd64 ARCH=x64 BUILDDIR=win32-x64 build-win32

platform-win32-ia32:
	@$(MAKE) --no-print-directory GOOS=windows GOARCH=386 ARCH=ia32 BUILDDIR=win32-ia32 build-win32

platform-win32-arm64:
	@$(MAKE) --no-print-directory GOOS=windows GOARCH=arm64 BUILDDIR=win32-arm64 build-win32

build-platform:
	@test -n "$(GOOS)" || (echo "The environment variable GOOS must be provided" && false)
	@test -n "$(GOARCH)" || (echo "The environment variable GOARCH must be provided" && false)
	@test -n "$(BUILDDIR)" || (echo "The environment variable BUILDDIR must be provided" && false)
	CGO_ENABLED=0 GOOS="$(GOOS)" GOARCH="$(GOARCH)" go build $(GO_FLAGS) -o "build/$(BUILDDIR)/package/bin/esbuild" ./esbuild_scss.go
	@$(MAKE) --no-print-directory GOOS="$(GOOS)" GOARCH="$(ARCH)" BUILDDIR="build/$(BUILDDIR)" download-nix
	tar -czf "build/esbuild-$(BUILDDIR)-$(VERSION).tgz" --directory "build/$(BUILDDIR)" package dart-sass

## Define the build targets for each platform.
platform-darwin-x64:
	@$(MAKE) --no-print-directory GOOS=darwin GOARCH=amd64 ARCH=x64 BUILDDIR=darwin-x64 build-platform

platform-darwin-arm64:
	@$(MAKE) --no-print-directory GOOS=darwin GOARCH=arm64 ARCH=arm64 BUILDDIR=darwin-arm64 build-platform

platform-freebsd-x64:
	@$(MAKE) --no-print-directory GOOS=freebsd GOARCH=amd64 ARCH=x64 BUILDDIR=freebsd-x64 build-platform

platform-freebsd-arm64:
	@$(MAKE) --no-print-directory GOOS=freebsd GOARCH=arm64 ARCH=arm64 BUILDDIR=freebsd-arm64 build-platform

platform-netbsd-x64:
	@$(MAKE) --no-print-directory GOOS=netbsd GOARCH=amd64 ARCH=x64 BUILDDIR=netbsd-x64 build-platform

platform-openbsd-x64:
	@$(MAKE) --no-print-directory GOOS=openbsd GOARCH=amd64 ARCH=x64 BUILDDIR=openbsd-x64 build-platform

platform-linux-x64:
	@$(MAKE) --no-print-directory GOOS=linux GOARCH=amd64 ARCH=x64 BUILDDIR=linux-x64 build-platform

platform-linux-ia32:
	@$(MAKE) --no-print-directory GOOS=linux GOARCH=386 ARCH=ia32 BUILDDIR=linux-ia32 build-platform

platform-linux-arm:
	@$(MAKE) --no-print-directory GOOS=linux GOARCH=arm ARCH=arm BUILDDIR=linux-arm build-platform

platform-linux-arm64:
	@$(MAKE) --no-print-directory GOOS=linux GOARCH=arm64 ARCH=arm64 BUILDDIR=linux-arm64 build-platform

platform-wasm:
	@$(MAKE) --no-print-directory GOOS=js GOARCH=wasm BUILDDIR=esbuild-wasm build-platform

clean:
	rm -rf build
	rm -rf *.zip
	rm -rf *.tar.gz
