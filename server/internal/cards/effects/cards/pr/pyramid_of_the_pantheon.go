package cardspr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Pyramid of the Pantheon — Artifact {1} (Amonkhet, S82 OLD batch 242, #5535):
//
//	"{2}, {T}: Add one mana of any color. Put a brick counter on this artifact.
//	 {T}: Add three mana of any one color. Activate only if there are three or
//	 more brick counters on this artifact."
//
// Two mana abilities (CR 605.1a). The first adds the mana, then its rider puts
// the brick counter on the Pyramid; the second is gated on the Pyramid's own
// brick counters, read as it is activated.
//
// No simplification.
func init() {
	const brick = "brick"
	effects.Register(effects.Spec{
		OracleID:     "efc66685-94cf-4d3f-82e0-9967c6ab2e5c",
		Name:         "Pyramid of the Pantheon",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{
			{
				Cost:     effects.ManaAbilityCost{Tap: true, Mana: "{2}"},
				Produced: "{W|U|B|R|G}",
				Label:    "{2}, {T}: Add one mana of any color. Put a brick counter on this artifact.",
				Rider: func(g *game.Game, _, source uuid.UUID) error {
					return g.AddCounterForEffect(source, brick, 1)
				},
			},
			{
				Cost:     effects.ManaAbilityCost{Tap: true},
				Produced: effects.OneColorOfAmount(3),
				Label:    "{T}: Add three mana of any one color. Activate only if there are three or more brick counters on this artifact.",
				Condition: func(g *game.Game, _, source uuid.UUID) bool {
					c, ok := g.LookupCardForEffect(source)
					return ok && c.Counters[brick] >= 3
				},
			},
		},
	})
}
