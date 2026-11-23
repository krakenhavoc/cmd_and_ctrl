package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bumi Bash — Sorcery {3}{R}:
//
//	"Choose one —
//	 • Bumi Bash deals damage equal to the number of lands you control
//	   to target creature.
//	 • Destroy target land creature or nonbasic land."
//
// Abrade's two-mode shape. The first mode counts the lands its
// controller controls as it resolves (an earthbent land creature is
// still a land and counts). The second mode's target is a permanent
// that is both a land and a creature, or any land without the basic
// supertype; a basic land that is not a creature is not a legal target.
// An illegal target makes the spell do nothing (CR 608.2b).
//
// No simplification. Avatar: The Last Airbender (#4459).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "629f247c-1e64-4c32-8892-0716ed469d1c",
		Name:         "Bumi Bash",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Modes: effects.ChooseOne(
			effects.Mode("Bumi Bash deals damage equal to the number of lands you control to target creature.",
				effects.TargetCreature("target creature")),
			effects.Mode("Destroy target land creature or nonbasic land.",
				effects.TargetPermanent("target land creature or nonbasic land",
					effects.And(effects.Land(), effects.Or(effects.Creature(), effects.NotABasicLandCard())))),
		),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			t, ok := effects.ModeTarget(ctx, 0)
			if !ok || t.Kind != game.TargetCard {
				return nil
			}
			target := t.ID
			switch {
			case ctx.HasMode(0):
				lands := effects.CountControlled(ctx.Game, ctx.Controller(), func(c game.Card) bool { return c.IsLand() })
				return effects.DealDamage{Source: ctx.Source(), Target: target, Amount: lands}.Apply(ctx)
			case ctx.HasMode(1):
				return effects.DestroyTarget{Target: target}.Apply(ctx)
			}
			return nil
		},
	})
}
