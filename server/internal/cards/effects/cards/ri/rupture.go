package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rupture — Sorcery {2}{R} (Nemesis, S82 OLD, #5618):
//
//	"Sacrifice a creature. Rupture deals damage equal to that creature's power
//	 to each creature without flying and each player."
//
// The sacrifice is on resolution, the controller's choice (Xathrid Demon's
// run), and the damage is the sacrificed creature's last-known power (CR
// 608.2h), dealt by Rupture to each creature without flying and to each player
// (Flamebreak's recipients), the caster included. With no creature to sacrifice
// nothing is dealt.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d2179a54-4b45-4b67-a31c-aa1a7bcfda2c",
		Name:         "Rupture",
		Completeness: effects.CompletenessFull,
		Purpose: game.Purpose{
			Answers: game.AnswerRemove,
			Sweep:   game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, AmountIsX: true},
		},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.OldB678Rupture(ctx.Game, item)
		},
	})
}
