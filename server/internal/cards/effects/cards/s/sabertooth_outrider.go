package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sabertooth Outrider — Creature — Human Warrior {3}{R}, 4/2 (Dragons of Tarkir,
// S82 older sets):
//
//	"Trample
//	 Formidable — Whenever this creature attacks, if creatures you control have
//	 total power 8 or greater, this creature gains first strike until end of
//	 turn."
//
// Surrak, the Hunt Caller's formidable on an attack trigger: an intervening "if"
// (CR 603.4) over the total current power of the creatures you control, read as
// the creature attacks and again as the ability resolves; the Outrider's own
// power counts. First strike is a layer 6 grant pinned to the Outrider until end
// of turn.
//
// No simplification.
func init() {
	const label = "Sabertooth Outrider — formidable: this creature gains first strike until end of turn"
	effects.Register(effects.Spec{
		OracleID:        "2a98f634-d739-47a7-98cb-b7f935cd65ef",
		Name:            "Sabertooth Outrider",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			effects.OldB324FormidableAttack(label, effects.ThisGainsKeywordUntilEndOfTurn("first strike", label)),
		},
	})
}
