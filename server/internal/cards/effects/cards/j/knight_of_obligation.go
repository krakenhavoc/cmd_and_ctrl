package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Knight of Obligation — Creature — Human Knight {3}{W}, 2/4 (Gatecrash, #5565):
//
//	"Vigilance
//	 Extort (Whenever you cast a spell, you may pay {W/B}. If you do, each
//	 opponent loses 1 life and you gain that much life.)"
//
// Vigilance is the engine keyword; extort is the shared Extort trigger
// (Basilica Screecher's).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "9051f778-7718-4e8d-88ff-ca2aee1b14e5",
		Name:            "Knight of Obligation",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered:       []game.TriggeredAbility{effects.Extort("Knight of Obligation")},
	})
}
