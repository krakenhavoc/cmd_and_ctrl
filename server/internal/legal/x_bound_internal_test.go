package legal

import (
	"math"
	"testing"
)

// #3895: Options.MaxX is a search bound on the bot, not a rule, but it
// never asks the engine about an X the engine refuses, and an open X
// range never offers one.
func TestMaxXAndTheOpenRangeStopAtTheEnginesBound(t *testing.T) {
	if got := (Options{MaxX: math.MaxInt}).withDefaults().MaxX; got != MaxAnnouncedX {
		t.Errorf("MaxX = %d, want %d", got, MaxAnnouncedX)
	}
	if got := (Options{}).withDefaults().MaxX; got != defaultMaxX {
		t.Errorf("default MaxX = %d, want %d", got, defaultMaxX)
	}
	v := openX(0, math.MaxInt)
	if v == nil || *v.Max != MaxAnnouncedX {
		t.Fatalf("openX(0, MaxInt) = %+v, want a range up to %d", v, MaxAnnouncedX)
	}
	if v := openX(MaxAnnouncedX+1, math.MaxInt); v != nil {
		t.Errorf("openX past the bound = %+v, want none", v)
	}
}
