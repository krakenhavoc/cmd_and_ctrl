package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dread Return — Sorcery {2}{B}{B}:
//
//	"Return target creature card from your graveyard to the
//	 battlefield.
//	 Flashback—Sacrifice three creatures. (You may cast this card from
//	 your graveyard for its flashback cost. Then exile it.)"
//
// The hand-cast half is an ordinary reanimation, the same shape
// Unburial Rites and Zombify use: the creature returns under its
// OWNER's control, which is the engine's reanimation default and the
// printed answer for "target creature card in YOUR graveyard".
//
// Sandbox simplification, declared: the flashback clause is NOT
// implemented. game.AlternativeCost has a mana cost, a life cost, and
// three card-naming components (ExileFromHand, ReturnToHand,
// ExileFromGraveyard) — nothing that pays with a SACRIFICE of the
// caster's own permanents. "Sacrifice three creatures" as a
// FLASHBACK cost (rather than an additional cost bolted onto the
// printed mana cost, which Spec.AdditionalCost already covers) needs
// a new field on AlternativeCost and the matching validate/pay/view
// plumbing — engine machinery this card file cannot add. Weaker than
// printed, never stronger: this card can only be cast from hand.
func init() {
	Register(Spec{
		OracleID:     "352b64d2-2ae5-44ee-a64f-94932ef545d3",
		Name:         "Dread Return",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"Flashback isn't implemented — this spell can only be cast from your hand for its printed mana cost, never later from the graveyard by sacrificing three creatures.",
		},
		Targets: TargetCardInGraveyard("target creature card in your graveyard", YouOwn(), Creature()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return returnFirstLegalGraveyardTargetToBattlefield(ctx.Game, item)
		},
	})
}
