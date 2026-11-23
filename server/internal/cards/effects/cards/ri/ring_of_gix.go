package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ring of Gix — Artifact, {3}:
//
//	"Echo {3} (At the beginning of your upkeep, if this came under your control since the beginning of your last upkeep, sacrifice it unless you pay its echo cost.)
//	 {1}, {T}: Tap target artifact, creature, or land."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e01fe9e9-a76c-4909-924c-a0cba403f975",
		Name:         "Ring of Gix",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, {T}: Tap target artifact, creature, or land.",
			Cost:    effects.Plus(effects.ManaCost("{1}"), effects.TapCost()),
			Targets: effects.TargetPermanent("target artifact, creature, or land", effects.Or(effects.Artifact(), effects.Creature(), effects.Land())),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				return effects.TapEach(ctx, effects.HoldTargetIDs(ctx))
			},
		}},
		Triggered: []game.TriggeredAbility{effects.Echo("Ring of Gix", "{3}")},
	})
}
