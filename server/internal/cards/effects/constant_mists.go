package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Constant Mists — "Buyback—Sacrifice a land. (You may sacrifice a
// land in addition to any other costs as you cast this spell. If you
// do, put this card into your hand as it resolves.) Prevent all
// combat damage that would be dealt this turn."
//
// The card #664 was opened for, and the reason the optional-cost flag
// went on the EXISTING AdditionalCost component rather than into a
// type of its own: strip the buyback and this is Fog with a worse
// mana cost, so there was never any point shipping it early.
//
// Its buyback is NON-MANA — sacrifice a land — which is the shape a
// mana-only design would have missed. BuybackSacrifice is the
// Sacrifice clause S21 shipped for Village Rites with Optional set,
// so the caster's pick is validated by the same validator, at
// announce, and the land is sacrificed with the spell already on the
// stack (CR 601.2h): a Blood Artist or a landfall-adjacent death
// payoff sees it and triggers ABOVE the Fog.
//
// The return to hand is the ENGINE's, not this file's. CR 702.27a
// replaces the resolving spell's destination, and
// routeStackCardToGraveyardLocked reads the paid record and routes to
// the owner's hand through the same stack-exit primitive flashback
// uses (ADR 0073 §6). Returning it here would be moving a card that
// is still on the stack.
//
// So the OnResolve is Fog's, verbatim, and that is the whole point:
// the mechanic is in the cost, not in the effect.
func init() {
	Register(Spec{
		OracleID:     "c850a29e-dc40-4ab6-89f1-d501a1a350d1",
		Name:         "Constant Mists",
		Completeness: CompletenessFull,
		OptionalCosts: []game.AdditionalCost{
			BuybackSacrifice("a land", Land()),
		},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return PreventAllCombatDamageThisTurn{
				Label: "Constant Mists: prevent combat damage",
			}.Apply(ctx)
		},
	})
}
