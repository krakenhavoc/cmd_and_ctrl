package playmat

import (
	"bytes"
	"encoding/binary"
	"errors"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"testing"
)

func solid(w, h int, c color.Color) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	r, g, b, a := c.RGBA()
	px := [4]uint8{uint8(r >> 8), uint8(g >> 8), uint8(b >> 8), uint8(a >> 8)}
	for i := 0; i < len(img.Pix); i += 4 {
		copy(img.Pix[i:i+4], px[:])
	}
	return img
}

func pngBytes(t *testing.T, img image.Image) []byte {
	t.Helper()
	var b bytes.Buffer
	if err := png.Encode(&b, img); err != nil {
		t.Fatal(err)
	}
	return b.Bytes()
}

func jpegBytes(t *testing.T, img image.Image) []byte {
	t.Helper()
	var b bytes.Buffer
	if err := jpeg.Encode(&b, img, nil); err != nil {
		t.Fatal(err)
	}
	return b.Bytes()
}

// withSegment inserts a marker segment right after SOI.
func withSegment(jpg []byte, marker byte, payload []byte) []byte {
	seg := []byte{0xFF, marker, byte((len(payload) + 2) >> 8), byte(len(payload) + 2)}
	seg = append(seg, payload...)
	out := append([]byte{}, jpg[:2]...)
	out = append(out, seg...)
	return append(out, jpg[2:]...)
}

// exifWithOrientation builds an APP1 payload with one IFD0 entry.
func exifWithOrientation(o uint16, gps string) []byte {
	var b bytes.Buffer
	b.WriteString("Exif\x00\x00")
	b.WriteString("MM")
	_ = binary.Write(&b, binary.BigEndian, uint16(42))
	_ = binary.Write(&b, binary.BigEndian, uint32(8))
	_ = binary.Write(&b, binary.BigEndian, uint16(1))
	_ = binary.Write(&b, binary.BigEndian, uint16(0x0112))
	_ = binary.Write(&b, binary.BigEndian, uint16(3))
	_ = binary.Write(&b, binary.BigEndian, uint32(1))
	_ = binary.Write(&b, binary.BigEndian, o)
	_ = binary.Write(&b, binary.BigEndian, uint16(0))
	_ = binary.Write(&b, binary.BigEndian, uint32(0))
	b.WriteString(gps)
	return b.Bytes()
}

func TestNormalizeAcceptsPNGAndJPEG(t *testing.T) {
	for name, data := range map[string][]byte{
		"png":  pngBytes(t, solid(40, 30, color.RGBA{200, 10, 10, 255})),
		"jpeg": jpegBytes(t, solid(40, 30, color.RGBA{10, 200, 10, 255})),
	} {
		n, err := Normalize(data)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if n.Width != 40 || n.Height != 30 {
			t.Errorf("%s: %dx%d, want 40x30", name, n.Width, n.Height)
		}
		if _, f, err := image.DecodeConfig(bytes.NewReader(n.JPEG)); err != nil || f != "jpeg" {
			t.Errorf("%s: stored bytes are %q (%v), want a jpeg", name, f, err)
		}
	}
}

func TestNormalizeRefusesWhatIsNotAnImage(t *testing.T) {
	for name, data := range map[string][]byte{
		"empty":      {},
		"text":       []byte("hello, this is not an image"),
		"html":       []byte("<html><script>alert(1)</script></html>"),
		"svg":        []byte(`<svg xmlns="http://www.w3.org/2000/svg" width="10" height="10"/>`),
		"gif":        []byte("GIF89a\x01\x00\x01\x00\x00\x00\x00;"),
		"truncated":  pngBytes(t, solid(50, 50, color.White))[:40],
		"png header": []byte("\x89PNG\r\n\x1a\n"),
	} {
		if _, err := Normalize(data); !errors.Is(err, ErrNotImage) {
			t.Errorf("%s: err = %v, want ErrNotImage", name, err)
		}
	}
}

func TestNormalizeRefusesAnOversizedFile(t *testing.T) {
	if _, err := Normalize(make([]byte, MaxUploadBytes+1)); !errors.Is(err, ErrTooLarge) {
		t.Fatalf("err = %v, want ErrTooLarge", err)
	}
}

// A decompression bomb: a tiny PNG that claims 100 000 x 100 000. The
// refusal must come from the header, before any pixel buffer exists.
func TestNormalizeRefusesBombDimensionsFromTheHeader(t *testing.T) {
	var b bytes.Buffer
	b.WriteString("\x89PNG\r\n\x1a\n")
	chunk := func(kind string, data []byte) {
		_ = binary.Write(&b, binary.BigEndian, uint32(len(data)))
		b.WriteString(kind)
		b.Write(data)
		_ = binary.Write(&b, binary.BigEndian, crc32Of(kind, data))
	}
	ihdr := make([]byte, 13)
	binary.BigEndian.PutUint32(ihdr[0:], 100_000)
	binary.BigEndian.PutUint32(ihdr[4:], 100_000)
	ihdr[8], ihdr[9] = 8, 2 // 8-bit RGB
	chunk("IHDR", ihdr)
	chunk("IEND", nil)
	if _, err := Normalize(b.Bytes()); !errors.Is(err, ErrTooManyPixels) {
		t.Fatalf("err = %v, want ErrTooManyPixels", err)
	}
}

func TestNormalizeShrinksTheLongEdge(t *testing.T) {
	n, err := Normalize(pngBytes(t, solid(3200, 1600, color.White)))
	if err != nil {
		t.Fatal(err)
	}
	if n.Width != MaxEdge || n.Height != MaxEdge/2 {
		t.Errorf("%dx%d, want %dx%d", n.Width, n.Height, MaxEdge, MaxEdge/2)
	}
}

func TestNormalizeNeverEnlarges(t *testing.T) {
	n, err := Normalize(pngBytes(t, solid(10, 7, color.White)))
	if err != nil {
		t.Fatal(err)
	}
	if n.Width != 10 || n.Height != 7 {
		t.Errorf("%dx%d, want 10x7", n.Width, n.Height)
	}
}

// Phone photos of a real playmat carry GPS in an APP1 EXIF segment.
// Re-encoding must drop it.
func TestNormalizeStripsEXIF(t *testing.T) {
	const marker = "SECRET-GPS-LAT-51.5007"
	src := withSegment(jpegBytes(t, solid(32, 32, color.White)), 0xE1, exifWithOrientation(1, marker))
	if !bytes.Contains(src, []byte(marker)) || !bytes.Contains(src, []byte("Exif")) {
		t.Fatal("test setup: the source JPEG has no EXIF segment")
	}
	n, err := Normalize(src)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(n.JPEG, []byte(marker)) || bytes.Contains(n.JPEG, []byte("Exif")) {
		t.Error("the stored JPEG still contains the EXIF segment")
	}
	// Every other metadata segment too: an APP1 XMP and a COM comment.
	src = withSegment(src, 0xFE, []byte("tracking-comment"))
	n, err = Normalize(src)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(n.JPEG, []byte("tracking-comment")) {
		t.Error("the stored JPEG still contains the COM segment")
	}
}

// Stripping EXIF without applying its orientation turns a portrait
// phone photo sideways.
func TestNormalizeAppliesEXIFOrientationBeforeStripping(t *testing.T) {
	// Orientation 6: the stored pixels are landscape, the photo is
	// portrait.
	src := withSegment(jpegBytes(t, solid(60, 20, color.White)), 0xE1, exifWithOrientation(6, ""))
	n, err := Normalize(src)
	if err != nil {
		t.Fatal(err)
	}
	if n.Width != 20 || n.Height != 60 {
		t.Errorf("%dx%d, want 20x60", n.Width, n.Height)
	}
	cfg, _, err := image.DecodeConfig(bytes.NewReader(n.JPEG))
	if err != nil || cfg.Width != 20 || cfg.Height != 60 {
		t.Errorf("stored image is %dx%d (%v), want 20x60", cfg.Width, cfg.Height, err)
	}
}

func TestOrientImageMovesAPixelWhereAViewerWouldShowIt(t *testing.T) {
	// A 3x2 image with a red pixel at its top-left.
	for o, want := range map[int]image.Point{
		2: {2, 0}, // mirror horizontal
		3: {2, 1}, // rotate 180
		4: {0, 1}, // mirror vertical
		5: {0, 0}, // transpose
		6: {1, 0}, // rotate 90 clockwise: top-left goes to top-right
		7: {1, 2}, // transverse
		8: {0, 2}, // rotate 90 counter-clockwise: top-left goes to bottom-left
	} {
		img := solid(3, 2, color.White)
		img.Set(0, 0, color.RGBA{255, 0, 0, 255})
		out := orientImage(img, o)
		if got := out.RGBAAt(want.X, want.Y); got.R != 255 || got.G != 0 {
			t.Errorf("orientation %d: red pixel is not at %v", o, want)
		}
	}
}

func TestJPEGOrientationOfAMalformedSegmentIsOne(t *testing.T) {
	base := jpegBytes(t, solid(8, 8, color.White))
	for name, payload := range map[string][]byte{
		"short":      []byte("Exif\x00\x00MM"),
		"bad magic":  []byte("Exif\x00\x00XX\x00\x2a\x00\x00\x00\x08\x00\x00"),
		"huge count": append([]byte("Exif\x00\x00MM\x00\x2a\x00\x00\x00\x08\xff\xff"), make([]byte, 4)...),
		"bad offset": []byte("Exif\x00\x00MM\x00\x2a\xff\xff\xff\xff\x00\x01\x00\x00\x00\x00"),
	} {
		if got := jpegOrientation(withSegment(base, 0xE1, payload)); got != 1 {
			t.Errorf("%s: orientation %d, want 1", name, got)
		}
	}
	if got := jpegOrientation([]byte("nope")); got != 1 {
		t.Errorf("not a jpeg: %d", got)
	}
}

func TestPNGAlphaIsFlattenedOntoWhite(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 16, 16)) // fully transparent
	n, err := Normalize(pngBytes(t, img))
	if err != nil {
		t.Fatal(err)
	}
	dec, err := jpeg.Decode(bytes.NewReader(n.JPEG))
	if err != nil {
		t.Fatal(err)
	}
	r, g, b, _ := dec.At(8, 8).RGBA()
	if r>>8 < 240 || g>>8 < 240 || b>>8 < 240 {
		t.Errorf("transparent pixel became (%d,%d,%d), want near white", r>>8, g>>8, b>>8)
	}
}

func TestReadLimited(t *testing.T) {
	if _, err := ReadLimited(bytes.NewReader(make([]byte, MaxUploadBytes))); err != nil {
		t.Errorf("exactly the cap: %v", err)
	}
	if _, err := ReadLimited(bytes.NewReader(make([]byte, MaxUploadBytes+1))); !errors.Is(err, ErrTooLarge) {
		t.Errorf("one over: %v, want ErrTooLarge", err)
	}
}
