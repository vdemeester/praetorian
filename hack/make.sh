#!/usr/bin/env bash
set -e

export PRAETORIAN_PKG='github.com/vdemeester/praetorian'
export SCRIPTDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
export MAKEDIR="$SCRIPTDIR/make"

# We're a nice, sexy, little shell script, and people might try to run us;
# but really, they shouldn't. We want to be in a container!
inContainer="AssumeSoInitially"
if [ "$PWD" != "/go/src/$PRAETORIAN_PKG" ]; then
    unset inContainer
fi

if [ -z "$inContainer" ]; then
    {
        echo "# WARNING! I don't seem to be running in a Docker container."
        echo "# The result of this command might be an incorrect build, and will not be"
        echo "# officially supported."
        echo "#"
        echo "# Try this instead: make all"
        echo "#"
    } >&2
fi

# List of bundles to create when no argument is passed
DEFAULT_BUNDLES=(
    validate-gofmt
    validate-govet
    validate-golint

    binary

    test-unit
    test-integration
)

TESTFLAGS+=" -test.timeout=10m"

# If $TESTFLAGS is set in the environment, it is passed as extra arguments to 'go test'.
# You can use this to select certain tests to run, eg.
#
#     TESTFLAGS='-test.run ^TestBuild$' ./hack/make.sh test-unit
#
# For integration-cli test, we use [gocheck](https://labix.org/gocheck), if you want
# to run certain tests on your local host, you should run with command:
#
#     TESTFLAGS='-check.f DockerSuite.TestBuild*' ./hack/make.sh binary test-integration-cli
#
go_test_dir() {
    dir=$1
    (
        echo '+ go test' $TESTFLAGS "${PRAETORIAN_PKG}${dir#.}"
        cd "$dir"
        export DEST="$ABS_DEST" # we're in a subshell, so this is safe -- our integration-cli tests need DEST, and "cd" screws it up
        go test $TESTFLAGS
    )
}

bundle() {
    local bundle="$1"; shift
    echo "---> Making bundle: $(basename "$bundle") (in $DEST)"
    source "${MAKEDIR}/$bundle" "$@"
}


main() {
    docker version
    echo $DOCKER_API_VERSION

    if [ $# -lt 1 ]; then
        bundles=(${DEFAULT_BUNDLES[@]})
    else
        bundles=($@)
    fi
    for bundle in ${bundles[@]}; do
        bundle "$bundle"
        echo
    done
}

main "$@"
