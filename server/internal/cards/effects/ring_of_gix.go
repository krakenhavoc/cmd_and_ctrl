package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ring of Gix — Artifact, {3}:
//
//	"Echo {3} (At the beginning of your upkeep, if this came under your control since the beginning of your last upkeep, sacrifice it unless you pay its echo cost.)
//	 {1}, {T}: Tap target artifact, creature, or land."
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e01fe9e9-a76c-4909-924c-a0cba403f975",
		Name:         "Ring of Gix",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{1}, {T}: Tap target artifact, creature, or land.",
			Cost:    Plus(ManaCost("{1}"), TapCost()),
			Targets: TargetPermanent("target artifact, creature, or land", Or(Artifact(), Creature(), Land())),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				return tapEach(ctx, holdTargetIDs(ctx))
			},
		}},
		Triggered: []game.TriggeredAbility{Echo("Ring of Gix", "{3}")},
	})
}
