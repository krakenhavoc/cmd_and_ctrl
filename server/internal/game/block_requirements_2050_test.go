package game

import (
	"testing"

	"github.com/google/uuid"
)

// TestLureExceptControllerBindsOnlyOthers — #2050: bindsBlocker spares
// the exempt player's creatures and binds everyone else's, judged by the
// controller now.
func TestLureExceptControllerBindsOnlyOthers(t *testing.T) {
	me, opp := uuid.New(), uuid.New()
	r := BlockRequirement{Kind: BlockRequirementLure, ExceptController: me}
	if r.bindsBlocker(&Card{Controller: me}) {
		t.Error("the exempt player's creature is bound")
	}
	if !r.bindsBlocker(&Card{Controller: opp}) {
		t.Error("another player's creature is not bound")
	}
	if r.bindsBlocker(nil) {
		t.Error("a nil blocker is bound")
	}
	if !(BlockRequirement{Kind: BlockRequirementLure}).bindsBlocker(&Card{Controller: me}) {
		t.Error("an ordinary Lure stopped binding")
	}
}
