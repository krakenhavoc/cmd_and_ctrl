package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bedevil — Instant {B}{B}{R} (EDHREC rank 483):
//
//	"Destroy target artifact, creature, or planeswalker."
//
// Rakdos's answer to almost everything that isn't an enchantment or
// a land. One target clause, three types OR-ed, the same shape
// Mortify and Putrefy use.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bceecc64-96f1-4e7b-8904-0aef90377764",
		Name:         "Bedevil",
		Completeness: effects.CompletenessFull,
		Targets: effects.TargetPermanent("target artifact, creature, or planeswalker",
			effects.Or(effects.Artifact(), effects.Creature(), effects.Planeswalker())),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			return effects.DestroyTarget{Target: item.Targets[0].ID}.Apply(ctx)
		},
	})
}
