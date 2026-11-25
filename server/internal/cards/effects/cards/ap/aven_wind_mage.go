package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Aven Wind Mage — Creature — Bird Wizard {2}{U}, 2/2 (Game Night 2019,
// #5515):
//
//	"Flying
//	 Whenever you cast an instant or sorcery spell, this creature gets +1/+1
//	 until end of turn."
//
// Frolicking Familiar's trigger: an instant or sorcery spell, with the pump
// on this creature as it resolves (it does nothing to a Mage that left and
// came back, CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ad5f9a18-89a5-4bcb-9cd7-5399ffd252d3",
		Name:            "Aven Wind Mage",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouCast(effects.Or(effects.Instant(), effects.Sorcery()),
				"Aven Wind Mage — gets +1/+1 until end of turn",
				effects.ThisGetsUntilEndOfTurn(1, 1, "Aven Wind Mage — +1/+1")),
		},
	})
}
