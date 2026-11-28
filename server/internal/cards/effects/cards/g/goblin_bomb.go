package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Goblin Bomb — Enchantment {1}{R} (Weatherlight, S82 older sets batch 780, #5629):
//
//	"At the beginning of your upkeep, you may flip a coin. If you win the flip, put a fuse counter on this enchantment.
//	 If you lose the flip, remove a fuse counter from this enchantment.
//	 Remove five fuse counters from this enchantment and sacrifice it: It deals 20 damage to target player or
//	 planeswalker."
//
// The "may" is asked as the trigger resolves (CR 603.5), then the coin is called and flipped (CR 705): a win adds a fuse
// counter, a loss removes one if there is one. The activated ability's cost is removing five fuse counters and sacrificing the
// Bomb (CR 602.2b); it targets a player or planeswalker and deals 20 damage, using the Bomb's last known information if
// it is gone (CR 608.2h).
//
// No simplification.
func init() {
	const label = "Goblin Bomb — you may flip a coin"
	effects.Register(effects.Spec{
		OracleID:     "061e2ef6-ce8d-4319-b525-96b9c2824001",
		Name:         "Goblin Bomb",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep(label+"; if you win, put a fuse counter on this enchantment, if you lose, remove one", effects.S82OldB780GoblinBombUpkeep(label)),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "Remove five fuse counters from this enchantment and sacrifice it: It deals 20 damage to target player or planeswalker.",
			Cost:    effects.Plus(effects.RemoveCountersFromThis("fuse", 5), effects.SacrificeThis()),
			Targets: effects.TargetPlayerOrPlaneswalker(),
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 20)),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					return effects.DealDamage{Source: ctx.Source(), Target: t.ID, Amount: 20}.Apply(ctx)
				}
				return nil
			},
		}},
	})
}
