package effects

import (
	"strconv"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// tor_b1_helpers.go — helpers for Torment, batch 1 (#5612).

// torB1FlashbackPayingLife is "Flashback—<mana>, Pay <life> life." (CR
// 702.34a): Acorn Harvest's {1}{G} and Crippling Fatigue's {1}{B}, both
// with 3 life. It is Flashback(mana) with AlternativeCost.Life set, the
// shape Deep Analysis spells out for its own price: the life is checked
// before anything is paid, so a caster with less life than that can't
// claim the offer at all (CR 118.3), and the constructor keeps the
// graveyard binding and the exile as the spell leaves the stack.
func torB1FlashbackPayingLife(mana string, life int) game.AlternativeCost {
	fb := Flashback(mana)
	fb.Label = "Flashback—" + mana + ", Pay " + strconv.Itoa(life) + " life"
	fb.Life = life
	return fb
}
