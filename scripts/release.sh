#!/usr/bin/env bash
#
# Writes dist/manifest.json for the binaries already built into dist/.
#
# The manifest is what a laptop in the field polls. Publish dist/ somewhere the
# laptop can reach over HTTPS and point UNLOCK_MANIFEST_URL at manifest.json; the
# sha256 in here is the only thing standing between the laptop and whatever the
# network hands it, so it is generated from the actual artifact every time.
set -euo pipefail

VERSION="${VERSION:?set VERSION}"
BINARY="${BINARY:-unlock}"
TARGET_OS="${TARGET_OS:-linux}"
TARGET_ARCH="${TARGET_ARCH:-amd64}"
# Where the artifacts will be served from, used to build the URLs.
BASE_URL="${BASE_URL:-https://example.invalid/unlock}"
NOTES="${NOTES:-}"

artifact="dist/${BINARY}-${TARGET_OS}-${TARGET_ARCH}"
[[ -f "$artifact" ]] || { echo "missing $artifact; run make dist first" >&2; exit 1; }

# sha256sum on Linux, shasum on macOS.
if command -v sha256sum >/dev/null 2>&1; then
  sum="$(sha256sum "$artifact" | cut -d' ' -f1)"
else
  sum="$(shasum -a 256 "$artifact" | cut -d' ' -f1)"
fi

size="$(wc -c < "$artifact" | tr -d ' ')"
# Strip a leading "v" so the manifest version compares cleanly against the
# version compiled into the binary.
clean_version="${VERSION#v}"

cat > dist/manifest.json <<EOF
{
  "version": "${clean_version}",
  "notes": "${NOTES}",
  "artifacts": {
    "${TARGET_OS}/${TARGET_ARCH}": {
      "url": "${BASE_URL}/${clean_version}/$(basename "$artifact")",
      "sha256": "${sum}",
      "size": ${size}
    }
  }
}
EOF

echo "wrote dist/manifest.json for ${clean_version}"
echo "  ${artifact}  ${sum}  ${size} bytes"
