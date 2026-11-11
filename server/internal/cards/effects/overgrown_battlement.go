package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "585f62dc-4461-42f1-a3a4-b19a1e550d2d",
		Name:            "Overgrown Battlement",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"defender"},
		ManaAbilities: []ManaAbility{{
			Cost: ManaAbilityCost{Tap: true},
			ProducedAmount: ProducedPerPermanent("G", func(c game.Card) bool {
				return c.IsCreature() && game.HasKeyword(&c, "defender")
			}),
			Label: "{T}: Add {G} for each creature you control with defender",
		}},
	})
}
