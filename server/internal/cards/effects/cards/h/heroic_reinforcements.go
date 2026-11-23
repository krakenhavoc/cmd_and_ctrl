package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Heroic Reinforcements — Sorcery {2}{R}{W} (Foundations, #4465):
//
//	"Create two 1/1 white Soldier creature tokens. Until end of turn,
//	 creatures you control get +1/+1 and gain haste. (They can attack and
//	 {T} this turn.)"
//
// The two Soldiers from the token table, then Goblin War Party's second
// bullet. The tokens are on the battlefield before the pump fixes its set
// of creatures (CR 611.2c), so they get +1/+1 and haste too; a creature
// that enters later this turn does not.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c442e8b0-a0a4-4839-8049-125b91b15c99",
		Name:         "Heroic Reinforcements",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			if err := (effects.CreateToken{Controller: ctx.Controller(), Template: effects.TokenCard("1/1 white Soldier"), N: 2}).Apply(ctx); err != nil {
				return err
			}
			yours := effects.And(effects.Creature(), effects.YouControl())
			if err := (effects.BoostUntilEOT{Match: yours, Power: 1, Toughness: 1, Label: "Heroic Reinforcements — +1/+1"}).Apply(ctx); err != nil {
				return err
			}
			return effects.GrantKeywordUntilEOT{Match: yours, Keywords: []string{"haste"}, Label: "Heroic Reinforcements — haste"}.Apply(ctx)
		},
	})
}
