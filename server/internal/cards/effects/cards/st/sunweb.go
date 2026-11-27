package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sunweb — Creature — Wall {3}{W}, 5/6 (Eighth Edition, #5607):
//
//	"Defender
//	 Flying
//	 This creature can't block creatures with power 2 or less."
//
// Power is read as blockers are declared (CR 509.1b, 208.1), through the
// layered view, so a pumped attacker may be blocked.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "832f96a4-0f4d-47bb-bc1c-3884e73feabc",
		Name:            "Sunweb",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender", "flying"},
		BlockRules: []game.BlockRule{
			effects.CantBlockAttackers(effects.OnSelf(), effects.OnMatching(effects.And(effects.Creature(), effects.PowerLE(2))),
				"it can't block creatures with power 2 or less"),
		},
	})
}
