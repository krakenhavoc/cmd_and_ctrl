package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "dbf9feb3-0751-47bd-a8dc-ed1c076cf8c7",
		Name:         "Broadside Barrage",
		Completeness: CompletenessFull,
		Targets:      TargetPermanent("target creature or planeswalker", Or(Creature(), Planeswalker())),
		Purpose:      game.Purpose{Draws: 1, Discards: 1, Targets: game.ForTargets(DamageToTarget(0, 5))},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			if err := (DealDamage{Source: item.SourceCardID, Target: item.Targets[0].ID, Amount: 5}).Apply(ctx); err != nil {
				return err
			}
			return lootOne(ctx.Game, item, 1)
		},
	})
}
