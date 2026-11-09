package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Quag Feast — {1}{B} Sorcery (Aetherdrift, #4464):
//
//	"Choose target creature, planeswalker, or Vehicle. Mill two cards, then
//	 destroy the chosen permanent if its mana value is less than or equal to
//	 the number of cards in your graveyard."
//
// The target is chosen on the stack (CR 115.1a); the mill is the first
// instruction and the graveyard is counted after it (CR 608.2c), so the
// destroy is the mill's continuation (MillToZone.ThenTail): a milled commander
// that waits on its owner's CR 903.9 answer is counted once it has landed. A
// Vehicle that is not a creature is a legal target. The mana value is read as
// the continuation runs, from the permanent's characteristics now (CR 202.3).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "78957246-4092-4f4a-80e3-142e24226962",
		Name:         "Quag Feast",
		Completeness: CompletenessFull,
		Targets: TargetPermanent("target creature, planeswalker, or Vehicle",
			Or(Creature(), Planeswalker(), HasSubtype("Vehicle"))),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if _, ok := ctx.ClauseTarget(0); !ok {
				return nil
			}
			return MillToZone{N: 2, ThenTail: ctx.ItemTail(quagFeastMilled)}.Apply(ctx)
		},
	})
}

// quagFeastMilled is the clause after the mill: destroy the chosen permanent if
// its mana value is at most the number of cards in the graveyard now. An
// on-disk key: never renamed, never reused.
var quagFeastMilled = CardTail("quag-feast/after-mill", func(ctx *Context, _ game.TailOutcome) error {
	t, ok := ctx.ClauseTarget(0)
	if !ok {
		return nil
	}
	c, ok := ctx.Game.LookupCardForEffect(t.ID)
	if !ok {
		return nil
	}
	mv, ok := ctx.Game.ManaValueForEffect(c)
	if !ok || mv > b31GraveyardSize(ctx.Game, ctx.Controller()) {
		return nil
	}
	return DestroyTarget{Target: t.ID}.Apply(ctx)
})
