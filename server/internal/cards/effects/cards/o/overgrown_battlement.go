package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Overgrown Battlement — Creature — Wall {1}{G}, 0/4:
//
//	"Defender
//	 {T}: Add {G} for each creature you control with defender."
//
// A mana ability with a counted output (Gaea's Cradle's shape). The count
// reads effective keywords, so a creature that has lost or gained defender
// is counted as it stands, and it includes the Battlement itself.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "585f62dc-4461-42f1-a3a4-b19a1e550d2d",
		Name:            "Overgrown Battlement",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender"},
		ManaAbilities: []effects.ManaAbility{{
			Cost: effects.ManaAbilityCost{Tap: true},
			ProducedAmount: effects.ProducedPerPermanent("G", func(c game.Card) bool {
				return c.IsCreature() && game.HasKeyword(&c, "defender")
			}),
			Label: "{T}: Add {G} for each creature you control with defender",
		}},
	})
}
