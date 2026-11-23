package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Broadside Barrage — Instant {1}{U}{R} (Aetherdrift, #4464):
//
//	"Broadside Barrage deals 5 damage to target creature or planeswalker.
//	 Draw a card, then discard a card."
//
// Magmatic Sinkhole's damage with Refute's loot after it. One target, so
// a target that has gone by resolution counters the spell and there is no
// loot (CR 608.2b). The loot is lootOne: the discard opens once the draw
// is done.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "dbf9feb3-0751-47bd-a8dc-ed1c076cf8c7",
		Name:         "Broadside Barrage",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("target creature or planeswalker", effects.Or(effects.Creature(), effects.Planeswalker())),
		Purpose:      game.Purpose{Draws: 1, Discards: 1, Targets: game.ForTargets(effects.DamageToTarget(0, 5))},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			if err := (effects.DealDamage{Source: item.SourceCardID, Target: item.Targets[0].ID, Amount: 5}).Apply(ctx); err != nil {
				return err
			}
			return effects.LootOne(ctx.Game, item, 1)
		},
	})
}
