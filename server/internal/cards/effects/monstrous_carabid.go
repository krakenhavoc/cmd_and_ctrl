package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Monstrous Carabid — Creature — Insect {3}{B}{R}, 4/4 (Alara Reborn,
// #5586):
//
//	"This creature attacks each combat if able.
//	 Cycling {B/R} ({B/R}, Discard this card: Draw a card.)"
//
// Bloodrock Cyclops' attack requirement (CR 508.1d) and the shared Cycling
// row at a hybrid cost (CR 702.29a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2e29025e-6954-4fbd-933f-81c2309c704e",
		Name:         "Monstrous Carabid",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{AttacksEachCombat()},
		Activated:    []ActivatedAbility{Cycling("{B/R}")},
	})
}
