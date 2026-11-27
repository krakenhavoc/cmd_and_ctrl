package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b665_helpers.go — S82 OLD batch 665 (Invasion enchantments):
// Fight or Flight's "only creatures in the pile of their choice can attack
// this turn" and Stand or Fall's "only creatures in the chosen piles can
// block this turn".

// oldB665NoCreature names no object on the battlefield: an empty chosen pile
// still bars every creature, and a ScopeCreaturesExceptObjects record must
// name at least one object (exceptObjectsScopeProblem). A fixed id, never a
// permanent's, so the sparing set matches nothing.
var oldB665NoCreature = uuid.MustParse("00000000-0000-0000-0000-00000000c665")

// OldB665PileRefs are the objects of `chosen` that are still on the
// battlefield, as the objects they are now (CR 400.7), or the one object
// that matches nothing when none is.
func OldB665PileRefs(g *game.Game, chosen []uuid.UUID) []game.ObjectRef {
	var refs []game.ObjectRef
	for _, id := range chosen {
		if c, ok := g.LookupCardForEffect(id); ok && g.Battlefield.Contains(id) {
			refs = append(refs, game.ObjectRef{ID: id, Epoch: c.ObjectEpoch})
		}
	}
	if len(refs) == 0 {
		refs = []game.ObjectRef{{ID: oldB665NoCreature}}
	}
	return refs
}

// OldB665OnlyPileCanAttackThisTurn is "only creatures in the pile can attack
// this turn": every creature but the pile's can't attack, read live (a
// creature that arrives later is covered, CR 611.2c).
func OldB665OnlyPileCanAttackThisTurn(ctx *Context, chosen []uuid.UUID, label string) {
	ctx.Game.RegisterScopedRuleEffectForEffect(ctx.Source(), game.ScopeCreaturesExceptObjects, ctx.Controller(),
		[]game.Mod{game.OnlyTheseCanAttackMod(OldB665PileRefs(ctx.Game, chosen))},
		DurationUntilEndOfTurn(ctx), label)
}

// OldB665OnlyPileCanBlockThisTurn is "only creatures in the chosen pile can
// block this turn" for one defending player: that player's creatures outside
// the pile can't block, read live.
func OldB665OnlyPileCanBlockThisTurn(ctx *Context, defender uuid.UUID, chosen []uuid.UUID, label string) {
	mod := game.AddRestrictionsMod(game.CantBlock)
	mod.Objects = OldB665PileRefs(ctx.Game, chosen)
	ctx.Game.RegisterScopedRuleEffectForEffect(ctx.Source(), game.ScopeYourCreaturesExceptObjects, defender,
		[]game.Mod{mod}, DurationUntilEndOfTurn(ctx), label)
}
