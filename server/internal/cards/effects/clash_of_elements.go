package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Clash of Elements — Instant {1}{U}{R}:
//
//	"Choose target nonland permanent. Its owner may put it on top of
//	 their library. If they do, Clash of Elements deals 2 damage to
//	 them. If they didn't put the card on top of their library, they
//	 put it on the bottom."
//
// The question goes to the permanent's OWNER, not the caster. "Yes"
// tucks it on top and, only if it actually reached a library, deals 2
// damage to that owner (a commander whose owner takes the command-zone
// offer instead, CR 903.9, is not damaged). "No" tucks it on the
// bottom. A token ceases to exist either way.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "78304a06-4c9a-4cdf-b67f-067d88762381",
		Name:         "Clash of Elements",
		Completeness: CompletenessFull,
		Targets:      TargetPermanent("target nonland permanent", Nonland()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			ts := ctx.LegalTargets()
			if len(ts) == 0 || ts[0].Kind != game.TargetCard {
				return nil
			}
			id := ts[0].ID
			c, ok := ctx.Game.LookupCardForEffect(id)
			if !ok {
				return nil
			}
			owner := c.Owner
			return MayChoice{
				Player:   owner,
				Question: "Clash of Elements — put " + c.Name + " on top of your library? (Yes: take 2 damage. No: it goes on the bottom.)",
				YesLabel: "Top (take 2 damage)",
				NoLabel:  "Bottom",
				Then:     clashOfElementsAnswer,
				Carry:    []uuid.UUID{id},
			}.Apply(ctx)
		},
	})
}

// clashOfElementsAnswer is the permanent's owner's answer: it goes on top
// of their library and they take 2 damage, or it goes on the bottom (the
// carried ID is the permanent). Keyed (#4291).
var clashOfElementsAnswer = MayYesNo("confirm/clash-of-elements-top-or-bottom",
	func(ctx *Context, r game.ConfirmAnswer) error {
		owner := r.Chooser
		return ctx.Game.TuckCardsToLibraryKeyedForEffect([]uuid.UUID{carryAt(r, 0)}, game.TuckOptions{}, withCarry(ctx.ItemTail(clashOfElementsTucked), owner))
	},
	func(ctx *Context, r game.ConfirmAnswer) error {
		return ctx.Game.TuckToLibraryForEffect(carryAt(r, 0), true)
	})

// clashOfElementsTucked deals 2 to the tucked card's owner (#4253 slice 9).
var clashOfElementsTucked = CardTail("clash-of-elements/damage-owner-if-tucked", func(ctx *Context, o game.TailOutcome) error {
	g, owner, tucked := ctx.Game, tailCarry0(o), len(o.IDs) == 1

	if !tucked {
		return nil
	}
	return DealDamage{Source: ctx.Item.SourceCardID, Target: owner, Amount: 2}.Apply(NewContext(g, ctx.Item))
})
