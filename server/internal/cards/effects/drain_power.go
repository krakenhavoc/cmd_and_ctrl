package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Drain Power — Sorcery {U}{U}:
//
//	"Target player activates a mana ability of each land they control.
//	 Then that player loses all unspent mana and you add the mana lost this
//	 way."
//
// #3952, ADR 0200: the target player activates a mana ability of each land
// they control, choosing which one and its colours (the 2004-10-04
// rulings), through ActivateAManaAbilityOfEachLand, then loses their whole
// pool, mana that was already in it included. You add mana of the same
// types and amounts. It is new mana Drain Power makes, so a restriction or
// spend rider the lost mana carried from the ability that made it does not
// come with it.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "0669172d-396b-4f5a-9703-129c5c849b55",
		Name:         "Drain Power",
		Completeness: CompletenessFull,
		Targets:      TargetPlayer("target player"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			target, ok := firstLegalPlayerTarget(ctx)
			if !ok {
				return nil
			}
			caster := ctx.Controller()
			return ActivateAManaAbilityOfEachLand{
				Player: target,
				Card:   "Drain Power",
				Then: func(ctx *Context, lost []game.ManaToken) error {
					return addManaLike(ctx, caster, lost)
				},
			}.Apply(ctx)
		},
	})
}
