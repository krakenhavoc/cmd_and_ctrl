package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sonorous Howlbonder — Creature — Human Warrior {1}{B/R}{B/R}, 2/2
// (Ikoria: Lair of Behemoths, #5512):
//
//	"Menace
//	 Each creature you control with menace can't be blocked except by three
//	 or more creatures."
//
// Rampaging Ceratops' MinBlockers over the creatures its controller
// controls that have menace, read through the layers at each declaration
// (CR 509.1b). The largest minimum wins, so a menace creature needs three.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "89645420-1a14-46bf-b27d-966553a95463",
		Name:            "Sonorous Howlbonder",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"menace"},
		BlockRules: []game.BlockRule{
			effects.MinBlockers(effects.OnMatching(effects.And(effects.Creature(), effects.YouControl(), effects.HasKeyword("menace"))), 3),
		},
	})
}
