#!/usr/bin/env bash
# Builds the app with the automation plugin and captures its window, headless.
#
#   scripts/preview.sh build                      # build once (bunx mygo build -debug, tag e2e)
#   scripts/preview.sh shot OUT.png [capture.py args...]
#
# shot starts the built app under Xvfb with the current environment (set
# BLH_DEMO=1, BLH_DEMO_STATE=live, ... before it), captures, and quits.
set -euo pipefail
cd "$(dirname "$0")/.."

case "${1:-}" in
build)
  GOFLAGS=-tags=e2e bunx mygo build -debug -skip-installer
  rm -rf .mygo/preview && mkdir -p .mygo/preview
  tar xzf build/linux-*/*.tar.gz -C .mygo/preview
  ;;
shot)
  out=$(realpath -m "$2"); shift 2
  bin=$(find .mygo/preview -maxdepth 2 -type f -perm -u+x | head -1)
  [ -n "$bin" ] || { echo "run: scripts/preview.sh build" >&2; exit 1; }
  sock=$(mktemp -u /tmp/blh-XXXX.sock)
  MYGO_AUTOMATION=$sock LIBGL_ALWAYS_SOFTWARE=1 \
    xvfb-run -a -s "-screen 0 1600x1000x24" "$bin" >/tmp/blh-preview.log 2>&1 &
  pid=$!
  trap 'pkill -P $pid 2>/dev/null; kill $pid 2>/dev/null; rm -f "$sock"' EXIT
  for _ in $(seq 1 40); do [ -S "$sock" ] && break; sleep 0.5; done
  python3 -I scripts/capture.py "$sock" "$out" "$@"
  ;;
*)
  sed -n 2,9p "$0"; exit 2
  ;;
esac
