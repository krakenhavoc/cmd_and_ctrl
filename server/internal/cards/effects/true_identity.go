package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// True Identity — Enchantment {1}{W} (Murders at Karlov Manor Commander,
// #5368):
//
//	"Whenever this enchantment or another permanent you control is turned
//	 face up, scry 1, then draw a card. This ability triggers only once
//	 each turn.
//	 Disguise {W}"
//
// Any permanent, not only a creature, counts; the enchantment itself is
// face up by the time the event is emitted (CR 708.8). The once-a-turn
// limit is TriggersOnlyOnceEachTurn.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "901d3027-4394-481e-9a0e-dec4513d9f88",
		Name:             "True Identity",
		Completeness:     CompletenessFull,
		Purpose:          game.Purpose{Scry: 1},
		AlternativeCosts: []game.AlternativeCost{Disguise("{W}")},
		Triggered: []game.TriggeredAbility{
			TriggersOnlyOnceEachTurn(
				On(game.EventTurnedFaceUp, APermanentYouControlWasTurnedFaceUp,
					"True Identity — scry 1, then draw a card", lembasScryThenDraw)),
		},
	})
}
