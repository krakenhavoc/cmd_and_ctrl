package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Megaton's Fate — Sorcery {5}{R}:
//
//	"Choose one —
//	 • Disarm — Destroy target artifact. Create four Treasure tokens.
//	 • Detonate — Megaton's Fate deals 8 damage to each creature. Each
//	   player gets four rad counters."
//
// #2042. Disarm's Treasures are the caster's, and they are made only if
// the spell resolves (an artifact gone in response leaves it with no
// legal target, CR 608.2b). Detonate's damage is one simultaneous
// instance; every player still in the game, the caster included, gets
// the counters.
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "09a688c1-21f6-4c7d-a6bd-aedc0d76a671",
		Name:         "Megaton's Fate",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeDoing("Disarm — Destroy target artifact. Create four Treasure tokens.",
				effects.TargetPermanent("target artifact", effects.Artifact()),
				func(item *game.StackItem, ctx *effects.Context, occ int) error {
					if err := effects.DestroyTheModesTarget(item, ctx, occ); err != nil {
						return err
					}
					return effects.CreateToken{Controller: ctx.Controller(), Template: effects.TreasureToken(), N: 4}.Apply(ctx)
				}),
			effects.ModeWithPurpose(effects.ModeDoing("Detonate — Megaton's Fate deals 8 damage to each creature. Each player gets four rad counters.",
				nil,
				func(_ *game.StackItem, ctx *effects.Context, _ int) error {
					if err := effects.DamageEachMatching(ctx, effects.Creature(), 8); err != nil {
						return err
					}
					return effects.EachPlayerGetsRadCounters(ctx.Game, ctx.Controller(), 4)
				}), game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 8}}),
		),
	})
}
