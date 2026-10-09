package playmat

// Tests for the best-size suggestion and the fit (ADR 0128 §11).

import (
	"bytes"
	"context"
	"errors"
	"image"
	"image/color"
	"image/jpeg"
	"testing"
)

func TestFitsMeansTheRightShapeAndEnoughPixels(t *testing.T) {
	cases := []struct {
		name string
		w, h int
		want bool
	}{
		{"the ideal itself", 2400, 1400, true},
		{"the ideal at the stored cap", 2560, 1493, true},
		{"just inside the shape tolerance, wide", 2500, 1400 + 1, true}, // 1.7843 vs 1.7143: +4.1%
		{"just outside the shape tolerance, wide", 2600, 1400, false},   // +8.3%
		{"just inside, narrow", 1700, 1000, true},                       // 1.70
		{"just outside, narrow", 1600, 1000, false},                     // 1.60 is -6.7%
		{"the right shape but small", 1200, 700, false},
		{"the right shape at the long-edge floor", 1600, 933, true},
		{"one pixel under the floor", 1599, 933, false},
		{"square", 2000, 2000, false},
		{"portrait", 1000, 2000, false},
		{"zero", 0, 0, false},
	}
	for _, c := range cases {
		if got := Fits(c.w, c.h); got != c.want {
			t.Errorf("%s: Fits(%d, %d) = %v, want %v", c.name, c.w, c.h, got, c.want)
		}
	}
}

func TestSuggestForALandscapeImageKeepsTheHeightAndCentresTheCrop(t *testing.T) {
	// 2560 x 1000 is much wider than 12:7: keep the whole height, trim
	// the sides. The crop is centred and has the ideal shape.
	s, ok := Suggest(2560, 1000)
	if !ok {
		t.Fatal("no suggestion for a wrong-shaped image")
	}
	if s.Crop.Height != 1000 || s.Crop.Width != 1714 || s.Crop.Y != 0 || s.Crop.X != (2560-1714)/2 {
		t.Errorf("crop = %+v", s.Crop)
	}
	// 1714 px of art cannot reach 2400: the result is the crop itself.
	if !s.Smaller || s.TargetWidth != 1714 || s.TargetHeight != 1000 {
		t.Errorf("target %dx%d smaller=%v, want 1714x1000 smaller=true (never upscale)", s.TargetWidth, s.TargetHeight, s.Smaller)
	}
}

func TestSuggestForAPortraitImageKeepsTheWidthAndCentresTheCrop(t *testing.T) {
	s, ok := Suggest(1000, 2000)
	if !ok {
		t.Fatal("no suggestion for a portrait image")
	}
	if s.Crop.Width != 1000 || s.Crop.Height != 583 || s.Crop.X != 0 || s.Crop.Y != (2000-583)/2 {
		t.Errorf("crop = %+v", s.Crop)
	}
	if !s.Smaller || s.TargetWidth != 1000 || s.TargetHeight != 583 {
		t.Errorf("target %dx%d smaller=%v", s.TargetWidth, s.TargetHeight, s.Smaller)
	}
}

func TestSuggestForALargeImageScalesDownToTheIdeal(t *testing.T) {
	s, ok := Suggest(2560, 1800) // taller than 12:7, plenty of pixels
	if !ok {
		t.Fatal("no suggestion")
	}
	if s.Crop.Width != 2560 || s.Crop.Height != 1493 {
		t.Errorf("crop = %+v", s.Crop)
	}
	if s.Smaller || s.TargetWidth != IdealWidth || s.TargetHeight != IdealHeight {
		t.Errorf("target %dx%d smaller=%v, want the ideal %dx%d", s.TargetWidth, s.TargetHeight, s.Smaller, IdealWidth, IdealHeight)
	}
}

func TestSuggestForASmallImageNeverUpscales(t *testing.T) {
	s, ok := Suggest(800, 800)
	if !ok {
		t.Fatal("no suggestion")
	}
	if s.TargetWidth > s.Crop.Width || s.TargetHeight > s.Crop.Height {
		t.Errorf("target %dx%d is larger than the crop %+v", s.TargetWidth, s.TargetHeight, s.Crop)
	}
	if !s.Smaller {
		t.Error("a small result must say so")
	}
}

func TestSuggestOffersNothingForAnImageOfTheRightShape(t *testing.T) {
	for _, d := range [][2]int{{2400, 1400}, {1200, 700}, {1600, 933}} {
		if s, ok := Suggest(d[0], d[1]); ok {
			t.Errorf("Suggest(%d, %d) = %+v, want none: a crop would change nothing", d[0], d[1], s)
		}
	}
}

func TestEverySuggestedCropIsTheIdealShapeAndInsideTheImage(t *testing.T) {
	for _, d := range [][2]int{{2560, 1000}, {1000, 2000}, {2560, 2560}, {301, 900}, {900, 301}, {5, 5}, {2560, 1800}, {1, 1}} {
		w, h := d[0], d[1]
		s, ok := Suggest(w, h)
		if !ok {
			continue
		}
		c := s.Crop
		if c.X < 0 || c.Y < 0 || c.X+c.Width > w || c.Y+c.Height > h || c.Width < 1 || c.Height < 1 {
			t.Errorf("%dx%d: crop %+v is outside the image", w, h, c)
		}
		if w >= 12 && h >= 7 && !aspectOK(c.Width, c.Height) {
			t.Errorf("%dx%d: crop %dx%d is not the ideal shape", w, h, c.Width, c.Height)
		}
	}
}

// bandedMat is a w x h image whose rows above split are red and the rest
// blue, so a test can tell WHICH rows a crop kept.
func bandedMat(t *testing.T, w, h, split int) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		c := color.RGBA{220, 20, 20, 255}
		if y >= split {
			c = color.RGBA{20, 20, 220, 255}
		}
		for x := 0; x < w; x++ {
			img.SetRGBA(x, y, c)
		}
	}
	return pngBytes(t, img)
}

func decodeStored(t *testing.T, s *Service, id string) image.Image {
	t.Helper()
	f, _, err := s.Open(id)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()
	img, err := jpeg.Decode(f)
	if err != nil {
		t.Fatal(err)
	}
	return img
}

func isRed(c color.Color) bool  { r, _, b, _ := c.RGBA(); return r>>8 > 150 && b>>8 < 100 }
func isBlue(c color.Color) bool { r, _, b, _ := c.RGBA(); return b>>8 > 150 && r>>8 < 100 }

func TestFitCropsScalesDownAndSwapsInANewFile(t *testing.T) {
	t.Parallel() // #2766: a full-size image under -race; own service, db and dir
	s, dir, user := newService(t, nil)
	ctx := context.Background()
	orig, err := s.SetFromBytes(ctx, user, 1, bandedMat(t, 2560, 1800, 900))
	if err != nil {
		t.Fatal(err)
	}
	if orig.Fits || orig.Suggestion == nil {
		t.Fatalf("a 2560x1800 mat should not fit: %+v", orig)
	}
	// Crop origin y = 0: rows 0..1492. Output row 700 is source row ~747, still red.
	got, err := s.Fit(ctx, user, 1, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	if got.Width != IdealWidth || got.Height != IdealHeight {
		t.Errorf("fitted size %dx%d, want %dx%d", got.Width, got.Height, IdealWidth, IdealHeight)
	}
	if got.ID == orig.ID || got.URL == orig.URL {
		t.Error("the fitted mat kept the old id, so caches would serve the old image")
	}
	if !got.Fits || got.Suggestion != nil {
		t.Errorf("the fitted mat still reports %+v", got)
	}
	img := decodeStored(t, s, got.ID)
	if b := img.Bounds(); b.Dx() != IdealWidth || b.Dy() != IdealHeight {
		t.Errorf("stored size %v", b)
	}
	if !isRed(img.At(1200, 600)) || !isBlue(img.At(1200, 1300)) {
		t.Error("crop at y=0 kept the wrong rows")
	}
	// The old file is gone and only the new one is on disk.
	if fs := filesIn(t, dir); len(fs) != 1 || fs[0] != got.ID+".jpg" {
		t.Errorf("files = %v, want only the fitted one", fs)
	}
	if _, _, err := s.Open(orig.ID); !errors.Is(err, ErrNotFound) {
		t.Errorf("the old file still opens: %v", err)
	}
	// Slot 1 was the active slot, so the table now shows the new file.
	if s.URL(user) != got.URL {
		t.Errorf("URL() = %q, want %q", s.URL(user), got.URL)
	}
	checkInvariant(t, s)
}

func TestFitHonoursWhereTheCropWindowSits(t *testing.T) {
	t.Parallel() // #2766: a full-size image under -race; own service, db and dir
	s, _, user := newService(t, nil)
	ctx := context.Background()
	if _, err := s.SetFromBytes(ctx, user, 1, bandedMat(t, 2560, 1800, 900)); err != nil {
		t.Fatal(err)
	}
	// The window is 1493 tall; the lowest origin is 1800 - 1493 = 307.
	got, err := s.Fit(ctx, user, 1, 0, 307)
	if err != nil {
		t.Fatal(err)
	}
	img := decodeStored(t, s, got.ID)
	// Rows 307..1799: output row 100 is source row ~407 (red); row 700
	// is ~1053 (blue). Origin 0 would have made row 700 red.
	if !isRed(img.At(1200, 100)) || !isBlue(img.At(1200, 700)) {
		t.Error("crop at y=307 did not keep the lower part of the art")
	}
}

func TestFitNeverUpscalesASmallImage(t *testing.T) {
	t.Parallel() // #2766: a full-size image under -race; own service, db and dir
	s, _, user := newService(t, nil)
	ctx := context.Background()
	if _, err := s.SetFromBytes(ctx, user, 2, mat(t, 1000, 2000, color.RGBA{9, 99, 9, 255})); err != nil {
		t.Fatal(err)
	}
	got, err := s.Fit(ctx, user, 2, 0, 700)
	if err != nil {
		t.Fatal(err)
	}
	if got.Width != 1000 || got.Height != 583 {
		t.Errorf("fitted %dx%d, want 1000x583 (the crop at its own resolution)", got.Width, got.Height)
	}
	if got.Width > IdealWidth || got.Height > IdealHeight {
		t.Error("the result is larger than the ideal")
	}
	checkInvariant(t, s)
}

func TestFitRefusesACropOutsideTheImage(t *testing.T) {
	t.Parallel() // #2766: a full-size image under -race; own service, db and dir
	s, dir, user := newService(t, nil)
	ctx := context.Background()
	orig, err := s.SetFromBytes(ctx, user, 1, bandedMat(t, 2560, 1800, 900))
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range [][2]int{{-1, 0}, {0, -1}, {0, 308}, {1, 0}, {2560, 0}, {0, 1800}, {1 << 30, 1 << 30}} {
		if _, err := s.Fit(ctx, user, 1, c[0], c[1]); !errors.Is(err, ErrBadCrop) {
			t.Errorf("Fit(%d, %d) = %v, want ErrBadCrop", c[0], c[1], err)
		}
	}
	// Nothing changed.
	if fs := filesIn(t, dir); len(fs) != 1 || fs[0] != orig.ID+".jpg" {
		t.Errorf("a refused fit changed the files: %v", fs)
	}
	if s.URL(user) != orig.URL {
		t.Error("a refused fit changed the playmat")
	}
}

func TestFitOnAnEmptySlotAndOnAnAlreadyFittingMat(t *testing.T) {
	t.Parallel() // #2766: a full-size image under -race; own service, db and dir
	s, _, user := newService(t, nil)
	ctx := context.Background()
	if _, err := s.Fit(ctx, user, 3, 0, 0); !errors.Is(err, ErrNoSlot) {
		t.Errorf("fitting an empty slot = %v, want ErrNoSlot", err)
	}
	if _, err := s.SetFromBytes(ctx, user, 1, mat(t, 2400, 1400, color.White)); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Fit(ctx, user, 1, 0, 0); !errors.Is(err, ErrAlreadyFits) {
		t.Errorf("fitting a mat that fits = %v, want ErrAlreadyFits", err)
	}
	// A fitted mat is the ideal shape, so a second fit is refused too.
	if _, err := s.SetFromBytes(ctx, user, 2, bandedMat(t, 2560, 1800, 900)); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Fit(ctx, user, 2, 0, 0); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Fit(ctx, user, 2, 0, 0); !errors.Is(err, ErrAlreadyFits) {
		t.Errorf("fitting twice = %v, want ErrAlreadyFits", err)
	}
}

func TestFitLeavesTheActivePointerAloneWhenTheSlotIsNotActive(t *testing.T) {
	t.Parallel() // #2766: a full-size image under -race; own service, db and dir
	s, _, user := newService(t, nil)
	ctx := context.Background()
	one, _ := s.SetFromBytes(ctx, user, 1, mat(t, 2400, 1400, color.White))
	if _, err := s.SetFromBytes(ctx, user, 2, bandedMat(t, 2560, 1800, 900)); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Fit(ctx, user, 2, 0, 0); err != nil {
		t.Fatal(err)
	}
	if s.URL(user) != one.URL {
		t.Errorf("URL() = %q, want slot 1's %q: fitting slot 2 must not switch", s.URL(user), one.URL)
	}
	checkInvariant(t, s)
}

func TestFitThatLosesARaceDropsItsFileAndSaysSo(t *testing.T) {
	t.Parallel() // #2766: a full-size image under -race; own service, db and dir
	s, dir, user := newService(t, nil)
	ctx := context.Background()
	orig, err := s.SetFromBytes(ctx, user, 1, bandedMat(t, 2560, 1800, 900))
	if err != nil {
		t.Fatal(err)
	}
	// Another request replaced the slot after the fit read it: the
	// crop was made from an image that is no longer there.
	repl, err := s.SetFromBytes(ctx, user, 1, mat(t, 64, 64, color.White))
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, solid(10, 10, color.White), nil); err != nil {
		t.Fatal(err)
	}
	if _, err := s.put(ctx, user, 1, buf.Bytes(), 10, 10, orig.ID); !errors.Is(err, ErrConflict) {
		t.Fatalf("put with a stale expectation = %v, want ErrConflict", err)
	}
	if fs := filesIn(t, dir); len(fs) != 1 || fs[0] != repl.ID+".jpg" {
		t.Errorf("files = %v, want only the replacement (the lost fit's file is deleted)", fs)
	}
	if s.URL(user) != repl.URL {
		t.Error("the lost fit changed the playmat")
	}
	checkInvariant(t, s)
}

func TestFitReadsHeaderSizeForARowWithNoStoredDimensions(t *testing.T) {
	t.Parallel() // #2766: a full-size image under -race; own service, db and dir
	// A row the migration backfilled has NULL width and height.
	s, _, user := newService(t, nil)
	ctx := context.Background()
	if _, err := s.SetFromBytes(ctx, user, 1, bandedMat(t, 2560, 1800, 900)); err != nil {
		t.Fatal(err)
	}
	if _, err := s.db.Exec(`UPDATE user_playmats SET width = NULL, height = NULL WHERE user_id = ?`, user.String()); err != nil {
		t.Fatal(err)
	}
	st, err := s.State(ctx, user)
	if err != nil || len(st.Slots) != 1 || st.Slots[0].Width != 2560 || st.Slots[0].Height != 1800 || st.Slots[0].Suggestion == nil {
		t.Fatalf("State = %+v, %v", st, err)
	}
	if _, err := s.Fit(ctx, user, 1, 0, 0); err != nil {
		t.Errorf("Fit on a backfilled row: %v", err)
	}
}
