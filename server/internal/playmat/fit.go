package playmat

// fit.go is ADR 0128 §11's best-size suggestion: the shape and size a
// playmat looks best at, whether a stored image already is that, and,
// when it is not, the crop that makes it so.
//
// Everything here is arithmetic on a width and a height. The image
// work (crop, scale, re-encode) is Service.Fit, and it is only ever
// asked for a rectangle this file would have offered.

import (
	"errors"
	"image"
	"image/draw"
	"math"

	xdraw "golang.org/x/image/draw"
)

const (
	// IdealWidth and IdealHeight are the size a playmat looks best at:
	// a paper playmat's 24 x 14 inches, which is 12:7, at 100 pixels
	// to the inch. ADR 0128 §11 records the measurement that chose
	// them: the battlefield area the board draws a seat's mat in is
	// within the tolerance of 12:7 at 1920x1080 and 2560x1440, so the
	// paper shape stands. Change the pair, and the ADR's table, together.
	IdealWidth  = 2400
	IdealHeight = 1400

	// AspectTolerance is how far from the ideal shape an image may be
	// and still "fit": 5% of the ideal width-to-height ratio, either way.
	AspectTolerance = 0.05

	// MinLongEdge is the long edge, in pixels, below which an image is
	// not "best" even in the right shape: it will look soft on a large
	// screen. Nothing is offered for it, because the fix is a bigger
	// image, and the server never enlarges one.
	MinLongEdge = 1600

	// MaxSlots is how many playmats an account keeps. Slots are 1 to
	// MaxSlots.
	MaxSlots = 3
)

// idealAspect is the ideal width over height.
const idealAspect = float64(IdealWidth) / float64(IdealHeight)

// Rect is a rectangle in a stored image's pixels.
type Rect struct {
	X      int `json:"x"`
	Y      int `json:"y"`
	Width  int `json:"width"`
	Height int `json:"height"`
}

// Suggestion is what "fit to best size" would do to a stored image: keep
// Crop, scale it down to TargetWidth x TargetHeight. It is offered only
// when the image's shape is off, so that a fit changes something.
type Suggestion struct {
	// TargetWidth and TargetHeight are the size the result will be:
	// the ideal size, or the crop's own size when the image is too small
	// to reach it. The server never scales up.
	TargetWidth  int `json:"target_width"`
	TargetHeight int `json:"target_height"`
	// Crop is the centred rectangle. The person may move it along the
	// axis being cropped (Service.Fit takes the origin).
	Crop Rect `json:"crop"`
	// Smaller is true when the result is smaller than the ideal size,
	// which the prompt warns about: the mat may look soft at the table.
	Smaller bool `json:"smaller"`
}

// Fits reports whether an image of w x h needs no work: its shape is
// within AspectTolerance of the ideal and its long edge is at least
// MinLongEdge.
func Fits(w, h int) bool {
	if w <= 0 || h <= 0 {
		return false
	}
	return aspectOK(w, h) && max(w, h) >= MinLongEdge
}

func aspectOK(w, h int) bool {
	return math.Abs(float64(w)/float64(h)/idealAspect-1) <= AspectTolerance
}

// cropSize is the largest ideal-shaped rectangle that fits inside w x h.
func cropSize(w, h int) (cw, ch int) {
	if float64(w)/float64(h) > idealAspect {
		// Wider than ideal: keep the full height, trim the sides.
		ch = h
		cw = int(math.Round(float64(h) * idealAspect))
	} else {
		cw = w
		ch = int(math.Round(float64(w) / idealAspect))
	}
	return min(max(cw, 1), w), min(max(ch, 1), h)
}

// Suggest returns the fit for an image of w x h. ok is false when none
// is worth offering: the image is already the right shape (whether or
// not it is large enough, since a crop would not change it), or it has
// no usable size.
func Suggest(w, h int) (s Suggestion, ok bool) {
	if w <= 0 || h <= 0 || aspectOK(w, h) {
		return Suggestion{}, false
	}
	cw, ch := cropSize(w, h)
	s = Suggestion{
		Crop: Rect{X: (w - cw) / 2, Y: (h - ch) / 2, Width: cw, Height: ch},
	}
	if cw >= IdealWidth {
		s.TargetWidth, s.TargetHeight = IdealWidth, IdealHeight
	} else {
		s.TargetWidth, s.TargetHeight = cw, ch
		s.Smaller = true
	}
	return s, true
}

// Errors from fitting. Each message is for the person who asked.
var (
	// ErrAlreadyFits means there is nothing to crop: the image is
	// already the ideal shape. A 409, not a silent success, so a client
	// that offers the action on a mat that does not need it learns so.
	ErrAlreadyFits = errors.New("this playmat is already the best shape")

	// ErrBadCrop means the rectangle is not one the server offered:
	// outside the image or at the wrong size. A 400.
	ErrBadCrop = errors.New("that crop does not lie inside the image")
)

// validCrop checks an origin against the image and returns the
// suggestion the fit will carry out.
func validCrop(w, h, x, y int) (Suggestion, error) {
	s, ok := Suggest(w, h)
	if !ok {
		return Suggestion{}, ErrAlreadyFits
	}
	if x < 0 || y < 0 || x > w-s.Crop.Width || y > h-s.Crop.Height {
		return Suggestion{}, ErrBadCrop
	}
	s.Crop.X, s.Crop.Y = x, y
	return s, nil
}

// cropAndScale returns the crop of src scaled to the suggestion's
// target, as an RGBA the caller encodes. It scales only down (or by
// nothing): TargetWidth never exceeds the crop's width.
func cropAndScale(src image.Image, s Suggestion) *image.RGBA {
	b := src.Bounds()
	r := image.Rect(b.Min.X+s.Crop.X, b.Min.Y+s.Crop.Y, b.Min.X+s.Crop.X+s.Crop.Width, b.Min.Y+s.Crop.Y+s.Crop.Height)
	dst := image.NewRGBA(image.Rect(0, 0, s.TargetWidth, s.TargetHeight))
	if s.TargetWidth == s.Crop.Width && s.TargetHeight == s.Crop.Height {
		draw.Draw(dst, dst.Bounds(), src, r.Min, draw.Src)
		return dst
	}
	xdraw.CatmullRom.Scale(dst, dst.Bounds(), src, r, xdraw.Src, nil)
	return dst
}
