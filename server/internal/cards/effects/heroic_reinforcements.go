package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "c442e8b0-a0a4-4839-8049-125b91b15c99",
		Name:         "Heroic Reinforcements",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if err := (CreateToken{Controller: ctx.Controller(), Template: TokenCard("1/1 white Soldier"), N: 2}).Apply(ctx); err != nil {
				return err
			}
			yours := And(Creature(), YouControl())
			if err := (BoostUntilEOT{Match: yours, Power: 1, Toughness: 1, Label: "Heroic Reinforcements — +1/+1"}).Apply(ctx); err != nil {
				return err
			}
			return GrantKeywordUntilEOT{Match: yours, Keywords: []string{"haste"}, Label: "Heroic Reinforcements — haste"}.Apply(ctx)
		},
	})
}
