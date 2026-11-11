package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Eumidian Hatchery — Land (Foundations Commander):
//
//	"{T}, Pay 1 life: Add {B}. Put a hatchling counter on this land.
//	 When this land is put into a graveyard from the battlefield, for each
//	 hatchling counter on it, create a 1/1 black Insect creature token
//	 with flying."
//
// Fiery Islet's {T}-and-life mana ability (ManaAbilityCost.Life) with the
// counter as its rider, in the same atomic resolution (the line is part of
// the mana ability, CR 605.3b). The death trigger counts the hatchling
// counters the land had as it left (CR 608.2h, Context.SourcePermanent),
// one token each; it fires for a destroyed, sacrificed or otherwise dead
// land alike.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5b6d933e-2830-4f5a-b244-a421aa9615dc",
		Name:         "Eumidian Hatchery",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true, Life: 1},
			Produced: "{B}",
			Label:    "{T}, Pay 1 life: Add {B}. Put a hatchling counter on this land.",
			Rider: func(g *game.Game, _, source uuid.UUID) error {
				return g.AddCounterForEffect(source, "hatchling", 1)
			},
		}},
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Eumidian Hatchery — create a 1/1 black Insect token with flying for each hatchling counter on it",
				Do(CreateTokensForEach{
					Template: TokenCard("1/1 black Insect with flying"),
					Name:     "the hatchling counters on this land",
					Count: func(ctx *Context) int {
						if info, ok := ctx.SourcePermanent(); ok {
							return info.Counters["hatchling"]
						}
						return 0
					},
				})),
		},
	})
}
