package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sarkhan's Dragonfire — Sorcery {3}{R}{R} (Core Set 2019, S82 OLD batch 198, #5527):
//
//	"Sarkhan's Dragonfire deals 3 damage to any target.
//	 Look at the top five cards of your library. You may reveal a red card from
//	 among them and put it into your hand. Put the rest on the bottom of your
//	 library in a random order."
//
// Roil Eruption's damage, then Skalla Wolf's look-at-five for a red card. The
// look happens even if the target has become illegal only when the spell
// still resolves (CR 608.2b): with its one target illegal the whole spell does
// nothing.
//
// No simplification.
func init() {
	look := effects.LookAtTopThenMayTakeToHand(5, effects.OfColor("R"), 1,
		"Sarkhan's Dragonfire — you may reveal a red card and put it into your hand")
	effects.Register(effects.Spec{
		OracleID:     "2a08d878-7816-4d0a-9532-4f2d3c8eee90",
		Name:         "Sarkhan's Dragonfire",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.DamageToTarget(0, 3)),
		Targets:      effects.TargetAny(),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			legal := ctx.LegalTargets()
			if len(legal) == 0 {
				return nil
			}
			if err := (effects.DealDamage{Target: legal[0].ID, Amount: 3, Source: item.SourceCardID}).Apply(ctx); err != nil {
				return err
			}
			return look(ctx.Game, item)
		},
	})
}
