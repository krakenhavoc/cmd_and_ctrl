package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Trinisphere — Artifact {3}:
//
//	"As long as this artifact is untapped, each spell that would cost
//	 less than three mana to cast costs three mana to cast.
//	 (Additional mana in the cost may be paid with any color of mana
//	 or colorless mana. For example, a spell that would cost {1}{B}
//	 to cast costs {2}{B} to cast instead.)"
//
// The catalog's only cost-SETTING effect, and the reason
// game.CostFloor is a third kind rather than a conditional increase.
// Two things about it are load-bearing:
//
//  1. It applies LAST. "Would cost less than three" has to mean the
//     price after every other increase and reduction on the board,
//     or a Goblin Electromancer could duck under the floor. The
//     engine runs the floor pass after the other two for exactly
//     this reason.
//  2. The shortfall is made up in GENERIC mana, which is what the
//     reminder text's own example says: {1}{B} becomes {2}{B}, not
//     {3}. The coloured requirement survives.
//
// "As long as this artifact is untapped" is an ordinary predicate on
// the modifier's own source — the cast path reads the Trinisphere's
// live battlefield state on every cast, so tapping it (Karn, the
// Great Creator's target, an opponent's Twiddle) switches the floor
// off with nothing to invalidate.
//
// Two Trinispheres behave as one, which is correct: the second floor
// finds a cost that already clears three and does nothing.
func init() {
	Register(Spec{
		OracleID:     "69d994f2-b8f6-425f-9655-977c2144d40c",
		Name:         "Trinisphere",
		Completeness: CompletenessFull,
		CostModifiers: []game.CostModifier{
			CostsAtLeast(3,
				"As long as this artifact is untapped, each spell that would cost less than three mana to cast costs three mana to cast.",
				SourceUntapped()),
		},
	})
}
