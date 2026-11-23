package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Basilica Guards — Creature — Human Soldier {2}{W}, 1/4 (Gatecrash, #5565):
//
//	"Defender
//	 Extort (Whenever you cast a spell, you may pay {W/B}. If you do, each
//	 opponent loses 1 life and you gain that much life.)"
//
// Defender is the engine keyword; extort is the shared Extort trigger
// (Basilica Screecher's).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "24d2e266-2dcb-40f1-a465-cfc63e9aa6ae",
		Name:            "Basilica Guards",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Triggered:       []game.TriggeredAbility{effects.Extort("Basilica Guards")},
	})
}
