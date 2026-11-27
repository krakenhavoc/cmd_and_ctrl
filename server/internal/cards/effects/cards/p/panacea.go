package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Panacea — Artifact {4} (Mercadian Masques, S82 OLD b681, #5619):
//
//	"{X}{X}, {T}: Prevent the next X damage that would be dealt to any target
//	 this turn."
//
// Battlefield Medic's charged shield (PreventNextDamage, CR 615.7) on any
// target, a creature, a player or a planeswalker; X is the one announced with
// the cost (CR 107.3k), so {X}{X} is twice X in mana. An X of 0 prevents
// nothing. A target gone by resolution gets no shield (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "08ee7879-3c9d-43c6-a44e-f81180f938fd",
		Name:         "Panacea",
		Completeness: effects.CompletenessFull,
		XMatters:     true,
		Activated: []effects.ActivatedAbility{{
			Label:   "{X}{X}, {T}: Prevent the next X damage that would be dealt to any target this turn.",
			Purpose: game.Purpose{Answers: game.AnswerPrevent},
			Cost:    effects.Plus(effects.ManaCost("{X}{X}"), effects.TapCost()),
			Targets: effects.TargetAny(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				ts := ctx.LegalTargets()
				if len(ts) == 0 || ctx.X() <= 0 {
					return nil
				}
				return effects.PreventNextDamage{Target: ts[0].ID, Amount: ctx.X(), Label: "Panacea: prevent the next X damage"}.Apply(ctx)
			},
		}},
	})
}
