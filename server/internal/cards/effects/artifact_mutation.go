package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Artifact Mutation — Instant {R}{G} (Invasion, #5616):
//
//	"Destroy target artifact. It can't be regenerated. Create X 1/1
//	 green Saproling creature tokens, where X is that artifact's mana
//	 value."
//
// destroyTargetThenSaprolingsForItsManaValue, also Aura Mutation's body.
// X is read as the spell resolves, just before the destruction, from the
// permanent as it is then (CR 608.2h), so a token artifact that copies
// nothing gives 0. The Saprolings are the caster's. They don't depend on
// the destruction: the text has no "if you do", so an indestructible
// artifact survives and the tokens are still made. An illegal target
// means the spell does nothing and makes no tokens (CR 608.2b), as the
// printed ruling says. The regeneration clause is CR 701.19c's.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e77e8768-3ea8-44f9-aba6-1e023e006946",
		Name:         "Artifact Mutation",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetPermanent("target artifact", Artifact()),
		OnResolve:    destroyTargetThenSaprolingsForItsManaValue(true),
	})
}

// destroyTargetThenSaprolingsForItsManaValue is "Destroy target
// <permanent>. Create X 1/1 green Saproling creature tokens, where X is
// that <permanent>'s mana value." as a spell's whole OnResolve (Artifact
// Mutation, Aura Mutation). X is the target's mana value just before it
// is destroyed; the tokens are the caster's and are made whether or not
// the destruction happened. With the target gone or illegal, nothing
// happens. noRegen is the printed "It can't be regenerated."
func destroyTargetThenSaprolingsForItsManaValue(noRegen bool) func(*game.StackItem, *Context) error {
	return func(_ *game.StackItem, ctx *Context) error {
		id, ok := b16FirstLegalTargetCard(ctx)
		if !ok {
			return nil
		}
		c, ok := ctx.Game.LookupCardForEffect(id)
		if !ok {
			return nil
		}
		// X is a value, not a gate on the destruction: the tokens come
		// whatever the destroy did (ADR 0013 §5m item 5).
		x := c.ManaValue()
		destroy := DestroyTarget{Target: id, CantBeRegenerated: noRegen}
		if x <= 0 {
			return destroy.Apply(ctx)
		}
		if err := destroy.Apply(ctx); err != nil {
			return err
		}
		return CreateToken{Controller: ctx.Controller(), Template: TokenCard("1/1 green Saproling"), N: x}.Apply(ctx)
	}
}
