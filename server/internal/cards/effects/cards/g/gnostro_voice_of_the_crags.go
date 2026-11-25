package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gnostro, Voice of the Crags — Legendary Creature — Chimera {1}{U}{R}{W},
// 3/3 (Commander Legends):
//
//	"{T}: Choose one. X is the number of spells you've cast this turn.
//	 • Scry X.
//	 • Gnostro deals X damage to target creature.
//	 • You gain X life."
//
// A modal activated ability (Bow of Nylea's shape) with one targeted bullet.
// X is read as the ability resolves, from the cast tally the engine keeps for
// each player (every spell the controller has cast this turn, countered ones
// included), so casting a spell in response changes it. With X at 0 each
// bullet does nothing.
//
// No simplification.
func init() {
	spells := func(ctx *effects.Context) int { return ctx.Game.CastTallyFor(ctx.Controller()).Total }
	effects.Register(effects.Spec{
		OracleID:     "bff55e1a-730c-4349-9ccb-18c97f46496d",
		Name:         "Gnostro, Voice of the Crags",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Choose one. X is the number of spells you've cast this turn. Scry X; or Gnostro deals X damage to target creature; or you gain X life.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.TapCost(),
			Modes: effects.ChooseOne(
				effects.ModeDoing("Scry X.", nil,
					func(_ *game.StackItem, ctx *effects.Context, _ int) error {
						x := spells(ctx)
						if x <= 0 {
							return nil
						}
						return effects.Scry{N: x}.Apply(ctx)
					}),
				effects.ModeDoing("Gnostro deals X damage to target creature.", effects.TargetCreature("target creature"),
					func(_ *game.StackItem, ctx *effects.Context, occ int) error {
						x := spells(ctx)
						for _, t := range ctx.ModeTargets(occ) {
							return effects.DealDamage{Source: ctx.Source(), Target: t.ID, Amount: x}.Apply(ctx)
						}
						return nil
					}),
				effects.ModeDoing("You gain X life.", nil,
					func(_ *game.StackItem, ctx *effects.Context, _ int) error {
						x := spells(ctx)
						if x <= 0 {
							return nil
						}
						return effects.GainLife{Amount: x}.Apply(ctx)
					}),
			),
		}},
	})
}
