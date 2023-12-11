## Platform Makefile
##

VERSION=0.19.5

platform-all:
	@$(MAKE) --no-print-directory -j4 \
		platform-darwin-arm64 \
		platform-darwin-x64 \
		platform-freebsd-arm64 \
		platform-freebsd-x64 \
		platform-linux-arm \
		platform-linux-arm64 \
		platform-linux-ia32 \
		platform-linux-loong64 \
		platform-linux-mips64el \
		platform-linux-ppc64 \
		platform-linux-riscv64 \
		platform-linux-s390x \
		platform-linux-x64 \
		platform-netbsd-x64 \
		platform-openbsd-x64 \
		platform-sunos-x64 \
		platform-wasm \
		platform-win32-arm64 \
		platform-win32-ia32 \
		platform-win32-x64

build-win32:
	CGO_ENABLED=0 GOOS="$(GOOS)" GOARCH="$(GOARCH)" go build $(GO_FLAGS) -o "build/$(BUILDDIR)/package/esbuild.exe" ./esbuild_scss.go
	tar -czf "build/$(BUILDDIR)-$(VERSION).tgz" --directory "build/$(BUILDDIR)" package

platform-win32-x64:
	@$(MAKE) --no-print-directory GOOS=windows GOARCH=amd64 BUILDDIR=win32-x64 build-win32

platform-win32-ia32:
	@$(MAKE) --no-print-directory GOOS=windows GOARCH=386 BUILDDIR=win32-ia32 build-win32

platform-win32-arm64:
	@$(MAKE) --no-print-directory GOOS=windows GOARCH=arm64 BUILDDIR=win32-arm64 build-win32

build-platform:
	@test -n "$(GOOS)" || (echo "The environment variable GOOS must be provided" && false)
	@test -n "$(GOARCH)" || (echo "The environment variable GOARCH must be provided" && false)
	@test -n "$(BUILDDIR)" || (echo "The environment variable BUILDDIR must be provided" && false)
	CGO_ENABLED=0 GOOS="$(GOOS)" GOARCH="$(GOARCH)" go build $(GO_FLAGS) -o "build/$(BUILDDIR)/package/bin/esbuild" ./esbuild_scss.go
	tar -czf "build/$(BUILDDIR)-$(VERSION).tgz" --directory "build/$(BUILDDIR)" package

## Define the build targets for each platform.
platform-darwin-x64:
	@$(MAKE) --no-print-directory GOOS=darwin GOARCH=amd64 BUILDDIR=darwin-x64 build-platform

platform-darwin-arm64:
	@$(MAKE) --no-print-directory GOOS=darwin GOARCH=arm64 BUILDDIR=darwin-arm64 build-platform

platform-freebsd-x64:
	@$(MAKE) --no-print-directory GOOS=freebsd GOARCH=amd64 BUILDDIR=freebsd-x64 build-platform

platform-freebsd-arm64:
	@$(MAKE) --no-print-directory GOOS=freebsd GOARCH=arm64 BUILDDIR=freebsd-arm64 build-platform

platform-netbsd-x64:
	@$(MAKE) --no-print-directory GOOS=netbsd GOARCH=amd64 BUILDDIR=netbsd-x64 build-platform

platform-openbsd-x64:
	@$(MAKE) --no-print-directory GOOS=openbsd GOARCH=amd64 BUILDDIR=openbsd-x64 build-platform

platform-linux-x64:
	@$(MAKE) --no-print-directory GOOS=linux GOARCH=amd64 BUILDDIR=linux-x64 build-platform

platform-linux-ia32:
	@$(MAKE) --no-print-directory GOOS=linux GOARCH=386 BUILDDIR=linux-ia32 build-platform

platform-linux-arm:
	@$(MAKE) --no-print-directory GOOS=linux GOARCH=arm BUILDDIR=linux-arm build-platform

platform-linux-arm64:
	@$(MAKE) --no-print-directory GOOS=linux GOARCH=arm64 BUILDDIR=linux-arm64 build-platform

platform-linux-loong64:
	@$(MAKE) --no-print-directory GOOS=linux GOARCH=loong64 BUILDDIR=linux-loong64 build-platform

platform-linux-mips64el:
	@$(MAKE) --no-print-directory GOOS=linux GOARCH=mips64le BUILDDIR=linux-mips64el build-platform

platform-linux-ppc64:
	@$(MAKE) --no-print-directory GOOS=linux GOARCH=ppc64le BUILDDIR=linux-ppc64 build-platform

platform-linux-riscv64:
	@$(MAKE) --no-print-directory GOOS=linux GOARCH=riscv64 BUILDDIR=linux-riscv64 build-platform

platform-linux-s390x:
	@$(MAKE) --no-print-directory GOOS=linux GOARCH=s390x BUILDDIR=linux-s390x build-platform

platform-sunos-x64:
	@$(MAKE) --no-print-directory GOOS=illumos GOARCH=amd64 BUILDDIR=sunos-x64 build-platform

platform-wasm:
	@$(MAKE) --no-print-directory GOOS=js GOARCH=wasm BUILDDIR=esbuild-wasm build-platform

clean:
	rm -rf build
