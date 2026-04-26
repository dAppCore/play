#!/usr/bin/env sh
set -eu

ROOT="${CORE_PLAY_ROOT:-tests/fixtures}"
BIN="${CORE_PLAY_BIN:-/tmp/core-play}"
OUT="${CORE_PLAY_SMOKE_OUT:-/tmp/core-play-smoke}"
HOME_ROOT="${CORE_PLAY_HOME:-/tmp/core-play-home}"

mkdir -p "$OUT"
mkdir -p "$HOME_ROOT"

GOWORK=off GOCACHE="${GOCACHE:-/tmp/core-play-gocache}" GOMODCACHE="${GOMODCACHE:-/tmp/core-play-gomodcache}" GOSUMDB="${GOSUMDB:-off}" \
	go build -tags engine_synthetic -o "$BIN" ./cmd/core-play

"$BIN" list --root "$ROOT" > "$OUT/list.txt"
cat "$OUT/list.txt"
CORE_PLAY_ROOT="$ROOT" "$BIN" verify sample-bundle > "$OUT/verify.txt"
cat "$OUT/verify.txt"
CORE_PLAY_ROOT="$ROOT" "$BIN" info sample-bundle > "$OUT/info.txt"
cat "$OUT/info.txt"
CORE_PLAY_ROOT="$ROOT" CORE_PLAY_HOME="$HOME_ROOT" "$BIN" sample-bundle > "$OUT/launch.txt"
cat "$OUT/launch.txt"
