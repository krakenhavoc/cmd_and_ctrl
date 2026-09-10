#!/usr/bin/env sh
# Rasterise the PWA icons from their SVG sources.
#
# The SVGs in client/public/icons are the source of truth; the PNGs
# next to them are build products that are committed, because the
# manifest references them by name and CI has no image toolchain.
# Re-run this after editing icon.svg / maskable.svg and commit the
# result.
#
# Renderer preference: rsvg-convert (librsvg) if present, else
# ImageMagick, else macOS QuickLook (qlmanage) -- which is what the
# committed PNGs were generated with, since this repo's dev machine
# has no other rasteriser. All three produce the same geometry; only
# the anti-aliasing differs slightly.
set -eu

cd "$(dirname "$0")/../public/icons"

render() { # render <src.svg> <size> <dest.png>
  src=$1
  size=$2
  dest=$3
  if command -v rsvg-convert >/dev/null 2>&1; then
    rsvg-convert -w "$size" -h "$size" -o "$dest" "$src"
  elif command -v magick >/dev/null 2>&1; then
    magick -background none "$src" -resize "${size}x${size}" "$dest"
  elif command -v qlmanage >/dev/null 2>&1; then
    tmp=$(mktemp -d)
    qlmanage -t -s "$size" -o "$tmp" "$src" >/dev/null 2>&1
    mv "$tmp/$(basename "$src").png" "$dest"
    rm -rf "$tmp"
  else
    echo "no SVG rasteriser found (install librsvg or imagemagick)" >&2
    exit 1
  fi
  echo "  $dest (${size}x${size})"
}

render icon.svg 192 icon-192.png
render icon.svg 512 icon-512.png
render maskable.svg 192 maskable-192.png
render maskable.svg 512 maskable-512.png
# iOS ignores the manifest and masks the corners itself, so the Apple
# touch icon comes from the full-bleed source, not the pre-rounded one.
render maskable.svg 180 apple-touch-icon.png
