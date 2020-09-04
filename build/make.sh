#!/usr/bin/env bash
set -e

# This script builds binary artifact from a checkout of formation source code.
#
# Requirements:
# - The current directory should be a checkout of the sds-formation source code.
#   Whatever version is checked out will be built.
# - The hash of the git commit will also be included in the formation
#   binary, with the suffix -dirty if the repository isn't clean.
# - The right way to call this script is to invoke "make" from
#   your checkout of the sds-formation repository.

set -o pipefail

export PKGNAME="xsky.com/sds-formation"
export SCRIPTDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
export MAKEDIR="$SCRIPTDIR/make"
export UTILSDIR="$SCRIPTDIR/utils"
export DEMON_CROSSPLATFORMS=(
    linux/amd64
)

echo

# List of directories containing source code
SRC_DIRS=$(find . -path './autogen/*' -prune -o \
     -name "*.go" -print0 | xargs -0n1 dirname | sort -u | sed 's,^\./,,g')

# List of bundles to create when no argument is passed
DEFAULT_BUNDLES=(
    validate-gofmt
    validate-test
    validate-vet
    test-unit
    binary
)

# get version tag
version=`git describe --tags --abbrev=0 | cut -d- -f1 | sed "s/^v//g"`
rpmVersion="$version"
# release is number of commits since the version tag
release=`git describe --tags | cut -d- -f2 | tr - _`

if [ "$version" = "$release" ]; then
    # no commits and release can't be empty
    VERSION="$version"
else
    VERSION="$version-$release"
fi

if command -v git &> /dev/null && git rev-parse &> /dev/null; then
    GITCOMMIT=$(git rev-parse --short HEAD)
    if [ -n "$(git status --porcelain --untracked-files=no)" ]; then
        GITCOMMIT="$GITCOMMIT-dirty"
    fi
    BUILDTIME=$(date -u)
else
    echo >&2 'error: .git directory missing not specified, Please either build with '
    echo >&2 'the .git directory accessible!'
    exit 1
fi

# Use these flags when compiling the tests and final binary
IAMSTATIC='true'
source "${MAKEDIR}/go-autogen"

HAVE_GO_TEST_COVER=
if \
    go help testflag | grep -- -cover > /dev/null \
    && go tool -n cover > /dev/null 2>&1 \
; then
    HAVE_GO_TEST_COVER=1
fi

# ORIG_BUILDFLAGS=( -a -tags "netgo static_build" -installsuffix netgo )
BUILDFLAGS=( $BUILDFLAGS )
# Test timeout.
: ${TIMEOUT:=60m}
TESTFLAGS+=" -timeout=${TIMEOUT}"

# If $TESTFLAGS is set in the environment, it is passed as extra arguments to 'go test'.
# You can use this to select certain tests to run, eg.
#
#     TESTFLAGS='-test.run ^TestBuild$' ./build/make.sh test-unit
#
# For integration-cli test, we use [gocheck](https://labix.org/gocheck), if you want
# to run certain tests on your local host, you should run with command:
#
#     TESTFLAGS='-check.f DockerSuite.TestBuild*' ./build/make.sh binary test-integration-cli
#
go_test_dir() {
    dir=$1
    coverpkg=$2
    testcover=()
    if [ "$HAVE_GO_TEST_COVER" ]; then
        # if our current go install has -cover, we want to use it :)
        mkdir -p "$DEST/coverprofiles"
        coverprofile="formation${dir#.}"
        coverprofile="$ABS_DEST/coverprofiles/${coverprofile//\//-}"
        testcover=( -cover -coverprofile "$coverprofile" $coverpkg )
    fi
    (
        echo '+ go test' $TESTFLAGS "${PKGNAME}${dir#.}"
        cd "$dir"
        export DEST="$ABS_DEST" # we're in a subshell, so this is safe -- our integration-cli tests need DEST, and "cd" screws it up
        test_env go test ${testcover[@]} -ldflags "$LDFLAGS" "${BUILDFLAGS[@]}" $TESTFLAGS
    )
}

test_env() {
    # use "env -i" to tightly control the environment variables that bleed into the tests
    env -i \
        DEST="$DEST" \
        GOPATH="$GOPATH" \
        HOME="$ABS_DEST/fake-HOME" \
        PATH="$PATH" \
        TEMP="$TEMP" \
        "$@"
}

replace_vars() {
    src=$1
    dst=$2
    if [[ -z $dst ]]; then
        dst=$src
    fi
    eval "cat <<EOF
$(<$src)
EOF" > $dst
}

# This helper function walks the current directory looking for directories
# holding certain files ($1 parameter), and prints their paths on standard
# output, one per line.
find_dirs() {
    find . -name "$1" -print0 | xargs -0n1 dirname | sort -u
}

hash_files() {
    while [ $# -gt 0 ]; do
        f="$1"
        shift
        dir="$(dirname "$f")"
        base="$(basename "$f")"
        for hashAlgo in md5 sha256; do
            if command -v "${hashAlgo}sum" &> /dev/null; then
                (
                    # subshell and cd so that we get output files like:
                    #   $HASH sds-formation-$VERSION
                    # instead of:
                    #   $HASH /go/src/github.com/.../$VERSION/binary/sds-formation-$VERSION
                    cd "$dir"
                    "${hashAlgo}sum" "$base" > "$base.$hashAlgo"
                )
            fi
        done
    done
}

init_rpm() {
    export TZ=UTC # make sure our "date" variables are UTC-based

    rpmName=$1
    pkgName=${rpmName}
    rpmVersion="${VERSION%%-*}"
    rpmRelease=1

    # get version tag
    version=`git describe --tags --abbrev=0 | cut -d- -f1 | sed "s/^v//g"`
    rpmVersion="$version"
    # release is number of commits since the version tag
    release=`git describe --tags | cut -d- -f2 | tr - _`

    if [ "$version" = "$release" ]; then
        # no commits and release can't be empty
        release=0
    fi

    hash=`git rev-parse HEAD | cut -c 1-8`
    release="$release.$hash"
    rpmRelease="$release"

    rpmPackager="$(awk -F ': ' '$1 == "Packager" { print $2; exit }' build/make/.build-rpm/${rpmName}.spec)"
    rpmDate="$(date +'%a %b %d %Y')"

    # if go-md2man is available, pre-generate the man pages
    ./man/md2man-all.sh -q || true
    # TODO decide if it's worth getting go-md2man in _each_ builder environment to avoid this

    os_version=$(cat /etc/centos-release)
    suite=$(echo $os_version | sed -n "s/CentOS Linux release \(.*\) (.*)/\1/p")
    if [[ -z $suite ]]; then
        suite=$(echo $os_version | sed -n "s/CentOS release \(.*\) (.*)/\1/p")
    fi
    if [[ -z $suite ]]; then
        echo "release version not found for $os_version"
        exit 1
    fi
    suite=$(echo $suite | cut -d'.' -f1)
    version="centos${suite}"

    rpmbuild_dir="$DEST/$version"
}

# a helper to provide ".exe" when it's appropriate
binary_extension() {
    if [ "$(go env GOOS)" = 'windows' ]; then
        echo -n '.exe'
    fi
}

bundle() {
    local bundle="$1"; shift
    echo "---> Making bundle: $(basename "$bundle") (in $DEST)"
    source "$SCRIPTDIR/make/$bundle" "$@"
}

main() {
    # We want this to fail if the bundles already exist and cannot be removed.
    # This is to avoid mixing bundles from different versions of the code.
    mkdir -p bundles
    if [ -e "bundles/$VERSION" ]; then
        echo "bundles/$VERSION already exists. Removing."
        rm -fr "bundles/$VERSION" && mkdir "bundles/$VERSION" || exit 1
        echo
    fi

    if [ "$(go env GOHOSTOS)" != 'windows' ]; then
        # Windows and symlinks don't get along well

        rm -f bundles/latest
        ln -s "$VERSION" bundles/latest
    fi

    if [ $# -lt 1 ]; then
        bundles=(${DEFAULT_BUNDLES[@]})
    else
        bundles=($@)
    fi
    for bundle in ${bundles[@]}; do
        export DEST="bundles/$VERSION/$(basename "$bundle")"
        mkdir -p "$DEST"
        ABS_DEST="$(cd "$DEST" && pwd -P)"
        bundle "$bundle"
        echo
    done
}

main "$@"
