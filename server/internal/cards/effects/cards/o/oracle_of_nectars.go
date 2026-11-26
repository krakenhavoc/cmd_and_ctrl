package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Oracle of Nectars — Creature — Elf Cleric {2}{G/W}, 2/2 (Shadowmoor, S82
// OLD b529, #5591):
//
//	"{X}, {T}: You gain X life."
//
// X is announced with the activation (CR 602.2b) and read as it resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f7ea8725-db0e-4ce5-ac77-7dfc2a4e9286",
		Name:         "Oracle of Nectars",
		Completeness: effects.CompletenessFull,
		XMatters:     true,
		Activated: []effects.ActivatedAbility{{
			Label:   "{X}, {T}: You gain X life.",
			Cost:    effects.Plus(effects.ManaCost("{X}"), effects.TapCost()),
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				return effects.GainLife{Player: ctx.Controller(), Amount: ctx.X()}.Apply(ctx)
			},
		}},
	})
}
