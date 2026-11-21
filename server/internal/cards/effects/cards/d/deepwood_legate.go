package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Deepwood Legate — Creature — Shade {3}{B}, 1/1 (Mercadian Masques, #5619):
//
//	"If an opponent controls a Forest and you control a Swamp, you may
//	 cast this spell without paying its mana cost.
//	 {B}: This creature gets +1/+1 until end of turn."
//
// The Legate cycle's free cast (FreeIfAnOpponentControlsAAndYouControlA):
// an OPPONENT controls a Forest and you control a Swamp, read off
// effective subtypes as you cast it. The pump is Fetid Horror's shade
// pump with the shared thisGetsUntilEndOfTurn body: no tap in the cost,
// so any number of activations, and the boost ends in the cleanup step
// (CR 514.2).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "414eb407-dd39-41d2-bf25-a7a77ff35cb4",
		Name:             "Deepwood Legate",
		Completeness:     effects.CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{effects.FreeIfAnOpponentControlsAAndYouControlA("Forest", "Swamp")},
		Activated: []effects.ActivatedAbility{{
			Label:   "{B}: This creature gets +1/+1 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Power: 1, Toughness: 1}},
			Cost:    effects.ManaCost("{B}"),
			Effect:  effects.ThisGetsUntilEndOfTurn(1, 1, "Deepwood Legate — +1/+1 until end of turn"),
		}},
	})
}
