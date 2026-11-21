package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Corrosive Gale — Sorcery {X}{G/P} (New Phyrexia, #5575):
//
//	"({G/P} can be paid with either {G} or 2 life.)
//	 Corrosive Gale deals X damage to each creature with flying."
//
// Needle Storm's flying sweep at X (ctx.X()). The Phyrexian symbol is
// paid by the engine at cast (CR 107.4f). It is a partial sweep: a
// creature without flying is spared.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6143073f-4110-4a37-bc0c-89fbda2fc839",
		Name:         "Corrosive Gale",
		Completeness: effects.CompletenessFull,
		XMatters:     true,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, AmountIsX: true, Partial: true}},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.DamageEachMatching(ctx, effects.And(effects.Creature(), effects.HasKeyword("flying")), ctx.X())
		},
	})
}
