#!/usr/bin/env sh
set -eu

ROOT="${CORE_PLAY_ROOT:-tests/fixtures}"
BIN="${CORE_PLAY_BIN:-/tmp/core-play}"
OUT="${CORE_PLAY_SMOKE_OUT:-/tmp/core-play-smoke}"
HOME_ROOT="${CORE_PLAY_HOME:-/tmp/core-play-home}"

mkdir -p "$OUT"
mkdir -p "$HOME_ROOT"

pass() {
	printf 'PASS %s\n' "$1"
}

require_contains() {
	file="$1"
	pattern="$2"
	if ! grep -F -q "$pattern" "$file"; then
		printf 'missing pattern %s in %s\n' "$pattern" "$file" >&2
		exit 1
	fi
}

hash_file() {
	if command -v sha256sum >/dev/null 2>&1; then
		sha256sum "$1" | awk '{print $1}'
	else
		shasum -a 256 "$1" | awk '{print $1}'
	fi
}

GOWORK=off GOCACHE="${GOCACHE:-/tmp/core-play-gocache}" GOMODCACHE="${GOMODCACHE:-/tmp/core-play-gomodcache}" GOSUMDB="${GOSUMDB:-off}" \
	go build -tags "engine_synthetic engine_dosbox engine_retroarch engine_scummvm" -o "$BIN" ./cmd/core-play

"$BIN" list --root "$ROOT" > "$OUT/list.txt"
cat "$OUT/list.txt"
require_contains "$OUT/list.txt" "sample-bundle"
pass "list"

CORE_PLAY_ROOT="$ROOT" "$BIN" verify sample-bundle > "$OUT/verify.txt"
cat "$OUT/verify.txt"
require_contains "$OUT/verify.txt" "OverallOK: true"
pass "verify"

CORE_PLAY_ROOT="$ROOT" "$BIN" info sample-bundle > "$OUT/info.txt"
cat "$OUT/info.txt"
require_contains "$OUT/info.txt" "name: sample-bundle"
pass "info"

CORE_PLAY_ROOT="$ROOT" CORE_PLAY_HOME="$HOME_ROOT" "$BIN" sample-bundle > "$OUT/launch.txt"
cat "$OUT/launch.txt"
require_contains "$OUT/launch.txt" "SYNTHETIC ENGINE OK"
pass "launch"

CORE_PLAY_ROOT="$ROOT" "$BIN" play shield-verify sample-bundle > "$OUT/shield.txt"
cat "$OUT/shield.txt"
require_contains "$OUT/shield.txt" "OverallOK: true"
require_contains "$OUT/shield.txt" "SBOM: [Y]"
require_contains "$OUT/shield.txt" "Code: [Y]"
require_contains "$OUT/shield.txt" "Content: [Y]"
require_contains "$OUT/shield.txt" "Threat: [Y]"
pass "shield-verify"

"$BIN" play engines > "$OUT/engines.txt"
cat "$OUT/engines.txt"
require_contains "$OUT/engines.txt" "dosbox"
require_contains "$OUT/engines.txt" "retroarch"
require_contains "$OUT/engines.txt" "scummvm"
pass "scummvm-list-platforms"

"$BIN" play list --root tests/fixtures/multi-bundle > "$OUT/catalogue.txt"
cat "$OUT/catalogue.txt"
require_contains "$OUT/catalogue.txt" "alpha"
require_contains "$OUT/catalogue.txt" "beta"
catalogue_count="$(grep -E '^(alpha|beta) ' "$OUT/catalogue.txt" | wc -l | awk '{print $1}')"
if [ "$catalogue_count" -ne 2 ]; then
	printf 'expected 2 catalogue rows, got %s\n' "$catalogue_count" >&2
	exit 1
fi
pass "catalogue"

DET_ROM="$OUT/det-rom.bin"
DET_ONE="$OUT/det-one"
DET_TWO="$OUT/det-two"
rm -rf "$DET_ONE" "$DET_TWO"
mkdir -p "$DET_ONE" "$DET_TWO"
printf 'DETERMINISTIC-ROM' > "$DET_ROM"
"$BIN" bundle --archive --root "$DET_ONE" --name det-test --title "Deterministic Test" --platform synthetic --engine synthetic --rom "$DET_ROM" > "$OUT/det-one.txt"
"$BIN" bundle --archive --root "$DET_TWO" --name det-test --title "Deterministic Test" --platform synthetic --engine synthetic --rom "$DET_ROM" > "$OUT/det-two.txt"
first_hash="$(hash_file "$DET_ONE/det-test.zip")"
second_hash="$(hash_file "$DET_TWO/det-test.zip")"
printf '%s\n%s\n' "$first_hash" "$second_hash" > "$OUT/deterministic.txt"
cat "$OUT/deterministic.txt"
if [ "$first_hash" != "$second_hash" ]; then
	printf 'deterministic bundle hashes differ\n' >&2
	exit 1
fi
pass "deterministic"

pass "all"
