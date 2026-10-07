// Package playmat is a signed-in person's playmats: up to three saved
// images per account, one of them active, that every player at a table
// sees behind that person's battlefield, like a mat on a paper table
// (ADR 0128).
//
// The package has these parts, each in its own file:
//
//   - process.go validates and normalises an image. The bytes a client
//     sends are never stored; what is stored is a fresh JPEG this
//     package encoded.
//   - fetch.go downloads a pasted URL once, behind an SSRF guard.
//   - store.go keeps the files on disk, keyed by uuid.
//   - service.go ties the file store to user_playmats (the three slots)
//     and users.playmat_id (the active one).
//   - fit.go is the best-size suggestion: what fits, and the crop that
//     makes an image fit (ADR 0128 §11).
//   - wash.go is the owner-set darkness under the cards (§10).
package playmat

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	_ "image/png" // registers the PNG decoder
	"io"

	"golang.org/x/image/draw"
	_ "golang.org/x/image/webp" // registers the WebP decoder
)

const (
	// MaxUploadBytes caps the bytes read from an upload or a fetched
	// URL. A 12 MP phone photo is 3 to 6 MB; 10 MB is a high-quality
	// original and still one request's worth of memory.
	MaxUploadBytes = 10 << 20

	// MaxPixels caps width times height BEFORE a full decode. A PNG of
	// a flat colour can be 100 KB and 100 000 000 pixels; decoding it
	// is 400 MB of RGBA. DecodeConfig reads only the header, so the
	// refusal costs nothing. 40 MP is above every phone camera and
	// below a decompression bomb.
	MaxPixels = 40_000_000

	// MaxEdge is the stored long edge. A battlefield area on a 4K
	// screen is under 2000 px wide, so more resolution is bytes every
	// other player downloads for nothing.
	MaxEdge = 2560

	// JPEGQuality is the stored JPEG quality.
	JPEGQuality = 85
)

// Errors a handler maps to a status. Each message is written for the
// person who uploaded the file.
var (
	// ErrNotImage means the bytes do not decode as PNG, JPEG or WebP,
	// whatever the Content-Type or the file name said.
	ErrNotImage = errors.New("that file is not a PNG, JPEG or WebP image")

	// ErrTooLarge means the file is over MaxUploadBytes.
	ErrTooLarge = errors.New("that image is larger than 10 MB")

	// ErrTooManyPixels means the dimensions are over MaxPixels.
	ErrTooManyPixels = errors.New("that image is larger than 40 megapixels")
)

// Normalized is a validated image, re-encoded for storage.
type Normalized struct {
	// JPEG is the bytes to store: a JPEG this package encoded, with no
	// EXIF, XMP or any other metadata segment.
	JPEG          []byte
	Width, Height int
}

// Normalize validates data by decoding it and returns the JPEG to
// store.
//
// The order is the safety argument:
//
//  1. Size, from the bytes in hand.
//  2. image.DecodeConfig: the format and dimensions from the header
//     alone, so a bomb is refused before any pixel buffer exists.
//  3. The pixel cap.
//  4. A full decode, which is the real "is this an image" test.
//  5. Re-encode. The JPEG written by image/jpeg carries only the
//     segments it writes itself, so EXIF (GPS from a phone photo of a
//     real playmat), XMP, ICC profiles and any payload appended to the
//     original are gone. The EXIF orientation is applied to the pixels
//     first, so a portrait phone photo does not come out sideways.
func Normalize(data []byte) (Normalized, error) {
	if len(data) > MaxUploadBytes {
		return Normalized{}, ErrTooLarge
	}
	cfg, format, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return Normalized{}, ErrNotImage
	}
	switch format {
	case "png", "jpeg", "webp":
	default:
		return Normalized{}, ErrNotImage
	}
	if cfg.Width <= 0 || cfg.Height <= 0 {
		return Normalized{}, ErrNotImage
	}
	// int64: a 32-bit int overflows on 65535 x 65535 in a 32-bit build,
	// and the check is the whole point of this function.
	if int64(cfg.Width)*int64(cfg.Height) > MaxPixels {
		return Normalized{}, ErrTooManyPixels
	}
	src, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return Normalized{}, ErrNotImage
	}

	orient := 1
	if format == "jpeg" {
		orient = jpegOrientation(data)
	}

	b := src.Bounds()
	w, h := scaledSize(b.Dx(), b.Dy(), MaxEdge)
	// Flatten onto white. JPEG has no alpha, and the scrim over the
	// mat is what keeps the board readable, not the mat's own
	// transparency.
	dst := image.NewRGBA(image.Rect(0, 0, w, h))
	draw.Draw(dst, dst.Bounds(), image.NewUniform(color.White), image.Point{}, draw.Src)
	if w == b.Dx() && h == b.Dy() {
		draw.Draw(dst, dst.Bounds(), src, b.Min, draw.Over)
	} else {
		draw.CatmullRom.Scale(dst, dst.Bounds(), src, b, draw.Over, nil)
	}

	var out image.Image = dst
	if orient > 1 {
		out = orientImage(dst, orient)
	}
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, out, &jpeg.Options{Quality: JPEGQuality}); err != nil {
		return Normalized{}, fmt.Errorf("playmat: encode: %w", err)
	}
	ob := out.Bounds()
	return Normalized{JPEG: buf.Bytes(), Width: ob.Dx(), Height: ob.Dy()}, nil
}

// ReadLimited reads r up to MaxUploadBytes, returning ErrTooLarge when
// there is more. It reads one byte past the cap to tell "exactly the
// cap" from "over it".
func ReadLimited(r io.Reader) ([]byte, error) {
	data, err := io.ReadAll(io.LimitReader(r, MaxUploadBytes+1))
	if err != nil {
		return nil, err
	}
	if len(data) > MaxUploadBytes {
		return nil, ErrTooLarge
	}
	return data, nil
}

// scaledSize shrinks (w, h) so the long edge is at most max, keeping
// the aspect ratio. It never enlarges.
func scaledSize(w, h, max int) (int, int) {
	long := w
	if h > long {
		long = h
	}
	if long <= max {
		return w, h
	}
	nw := (w*max + long/2) / long
	nh := (h*max + long/2) / long
	if nw < 1 {
		nw = 1
	}
	if nh < 1 {
		nh = 1
	}
	return nw, nh
}

// jpegOrientation reads the EXIF Orientation tag (1 to 8) from a JPEG,
// or 1 when there is none or the segment is malformed. It walks the
// marker segments up to the start of scan and looks only at APP1
// "Exif" data; nothing it reads is stored.
func jpegOrientation(data []byte) int {
	if len(data) < 4 || data[0] != 0xFF || data[1] != 0xD8 {
		return 1
	}
	i := 2
	for i+4 <= len(data) {
		if data[i] != 0xFF {
			return 1
		}
		marker := data[i+1]
		if marker == 0xFF { // fill byte
			i++
			continue
		}
		if marker == 0xDA || marker == 0xD9 { // start of scan, end of image
			return 1
		}
		if marker >= 0xD0 && marker <= 0xD8 || marker == 0x01 { // standalone markers
			i += 2
			continue
		}
		segLen := int(data[i+2])<<8 | int(data[i+3])
		if segLen < 2 || i+2+segLen > len(data) {
			return 1
		}
		if marker == 0xE1 {
			if o := exifOrientation(data[i+4 : i+2+segLen]); o != 0 {
				return o
			}
		}
		i += 2 + segLen
	}
	return 1
}

// exifOrientation reads the Orientation tag out of one APP1 payload.
// 0 means this segment holds none.
func exifOrientation(seg []byte) int {
	if len(seg) < 14 || string(seg[:6]) != "Exif\x00\x00" {
		return 0
	}
	t := seg[6:]
	var u16 func([]byte) int
	var u32 func([]byte) int
	switch string(t[:2]) {
	case "II":
		u16 = func(b []byte) int { return int(b[0]) | int(b[1])<<8 }
		u32 = func(b []byte) int { return int(b[0]) | int(b[1])<<8 | int(b[2])<<16 | int(b[3])<<24 }
	case "MM":
		u16 = func(b []byte) int { return int(b[0])<<8 | int(b[1]) }
		u32 = func(b []byte) int { return int(b[0])<<24 | int(b[1])<<16 | int(b[2])<<8 | int(b[3]) }
	default:
		return 0
	}
	if u16(t[2:4]) != 42 {
		return 0
	}
	off := u32(t[4:8])
	if off < 8 || off+2 > len(t) {
		return 0
	}
	n := u16(t[off : off+2])
	for k := 0; k < n; k++ {
		e := off + 2 + k*12
		if e+12 > len(t) {
			return 0
		}
		if u16(t[e:e+2]) == 0x0112 { // Orientation
			v := u16(t[e+8 : e+10])
			if v >= 1 && v <= 8 {
				return v
			}
			return 0
		}
	}
	return 0
}

// orientImage applies an EXIF orientation (2 to 8) to img, returning a
// new image with the pixels where a viewer that honours the tag would
// show them.
func orientImage(img *image.RGBA, o int) *image.RGBA {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	nw, nh := w, h
	if o >= 5 {
		nw, nh = h, w
	}
	out := image.NewRGBA(image.Rect(0, 0, nw, nh))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			var dx, dy int
			switch o {
			case 2: // mirror horizontal
				dx, dy = w-1-x, y
			case 3: // rotate 180
				dx, dy = w-1-x, h-1-y
			case 4: // mirror vertical
				dx, dy = x, h-1-y
			case 5: // transpose
				dx, dy = y, x
			case 6: // rotate 90 clockwise
				dx, dy = h-1-y, x
			case 7: // transverse
				dx, dy = h-1-y, w-1-x
			case 8: // rotate 90 counter-clockwise
				dx, dy = y, w-1-x
			default:
				dx, dy = x, y
			}
			si := img.PixOffset(x, y)
			di := out.PixOffset(dx, dy)
			copy(out.Pix[di:di+4], img.Pix[si:si+4])
		}
	}
	return out
}
