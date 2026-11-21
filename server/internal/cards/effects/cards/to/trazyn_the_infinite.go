package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Trazyn the Infinite — Legendary Artifact Creature — Necron {4}{B}{B},
// 4/6:
//
//	"Deathtouch
//	 Prismatic Gallery — As long as Trazyn is on the battlefield, it has
//	 all activated abilities of all artifact cards in your graveyard."
//
// A layer-6 borrow over the controller's graveyard (ADR 0155), the
// artifact cards' rows read from their catalog entries. "Prismatic
// Gallery" is an ability word and carries no rules of its own.
//
// No simplification.
func init() {
	const clause = "As long as Trazyn is on the battlefield, it has all activated abilities of all artifact cards in your graveyard."
	effects.Register(effects.Spec{
		OracleID:        "1ef3b48a-55ef-422e-ac7f-e1c1057eec2c",
		Name:            "Trazyn the Infinite",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Static: []game.StaticAbility{effects.HasActivatedAbilitiesOf(game.AbilityBorrow{
			Label:   clause,
			Lenders: effects.GraveyardLenders(effects.YourGraveyard, effects.ByController(effects.Artifact())),
		})},
	})
}
