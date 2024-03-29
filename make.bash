set -e

unset GOBIN # Issue 14340

if [ ! -f run.bash ]; then
    echo 'make.bash must be run from $GOROOT/src' 1>&2
    exit 1
fi


# Test for Windows.
case "$(uname)" in
*MINGW* | *WIN32* | *CYGWIN*)
    echo 'ERROR: Do not use make.bash to build on Windows.'
    echo 'Use make.bat instead.'
    echo
    exit 1
    ;;
esac

# Test for bad ld.
if ld --version 2>&1 | grep 'gold.* 2\.20' >/dev/null; then
    echo 'ERROR: Your system has gold 2.20 installed.'
    echo 'This version is shipped by Ubuntu even though'
    echo 'it is known not to work on Ubuntu.'
    echo 'Binaries built with this linker are likely to fail in mysterious ways.'
    echo
    echo 'Run sudo apt-get remove binutils-gold.'
    echo
    exit 1
fi

rm -f ./runtime/runtime_defs.go

echo '##### Building Go bootstrap tool.'
echo cmd/dist
export GOROOT="$(cd .. && pwd)"
echo "goroot="$GOROOT
GOROOT_BOOTSTRAP=${GOROOT_BOOTSTRAP:-$HOME/go1.4}
echo "GOROOT_BOOTSTRAP="$GOROOT_BOOTSTRAP

if [ ! -x "$GOROOT_BOOTSTRAP/bin/go" ]; then
    echo "ERROR: Cannot find $GOROOT_BOOTSTRAP/bin/go." >&2
    echo "Set \$GOROOT_BOOTSTRAP to a working Go tree >= Go 1.4." >&2
    exit 1
fi
if [ "$GOROOT_BOOTSTRAP" = "$GOROOT" ]; then
    echo "ERROR: \$GOROOT_BOOTSTRAP must not be set to \$GOROOT" >&2
    echo "Set \$GOROOT_BOOTSTRAP to a working Go tree >= Go 1.4." >&2
    exit 1
fi

rm -f cmd/dist/dist
echo GOROOT="$GOROOT_BOOTSTRAP" GOOS="" GOARCH="" "$GOROOT_BOOTSTRAP/bin/go" build -o cmd/dist/dist ./cmd/dist
GOROOT="$GOROOT_BOOTSTRAP" GOOS="" GOARCH="" "$GOROOT_BOOTSTRAP/bin/go" build -o cmd/dist/dist ./cmd/dist



# -e doesn't propagate out of eval, so check success by hand.
eval $(./cmd/dist/dist env -p || echo FAIL=true)
if [ "$FAIL" = true ]; then
    exit 1
fi

echo

if [ "$1" = "--dist-tool" ]; then
    # Stop after building dist tool.
    mkdir -p "$GOTOOLDIR"
    if [ "$2" != "" ]; then
        cp cmd/dist/dist "$2"
    fi
    mv cmd/dist/dist "$GOTOOLDIR"/dist
    exit 0
fi



buildall="-a"
if [ "$1" = "--no-clean" ]; then
    buildall=""
    shift
fi
echo ./cmd/dist/dist bootstrap $buildall $GO_DISTFLAGS -v # builds go_bootstrap
./cmd/dist/dist bootstrap $buildall $GO_DISTFLAGS -v # builds go_bootstrap
