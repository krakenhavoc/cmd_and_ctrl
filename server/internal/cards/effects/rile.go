package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rile — {R} Sorcery:
//
//	"Rile deals 1 damage to target creature you control. That creature
//	 gains trample until end of turn.
//	 Draw a card."
//
// The damage is dealt first and the trample is granted whether or not
// the creature survived (a dead creature is simply not there to
// receive it). The draw does not depend on the target: if the target
// is illegal on resolution the whole spell fizzles (CR 608.2b), since
// it has only the one target.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0c4e0e46-77f1-4ff6-afbc-1a67feceedfb",
		Name:         "Rile",
		Completeness: CompletenessFull,
		Purpose: game.Purpose{
			Draws: 1,
			Targets: game.ForTargets(game.TargetPurpose{
				Slot: 0, Damage: 1, UntilEndOfTurn: game.TargetBenefit,
				Pump: &game.Pump{Keywords: []string{"trample"}},
			}),
		},
		Targets: TargetCreature("target creature you control", YouControl()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok {
				return nil
			}
			if err := (DealDamage{Source: ctx.Source(), Target: t.ID, Amount: 1}).Apply(ctx); err != nil {
				return err
			}
			if err := (GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"trample"}, Label: "Rile — trample"}).Apply(ctx); err != nil {
				return err
			}
			return DrawCards{N: 1}.Apply(ctx)
		},
	})
}
