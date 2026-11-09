package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Maelstrom — Plane — Alara (Planechase Anthology, #4870):
//
//	"When you planeswalk to The Maelstrom and at the beginning of your
//	 upkeep, you may reveal the top card of your library. If a permanent
//	 card is revealed this way, you may put it onto the battlefield. If you
//	 revealed a card but didn't put it onto the battlefield, put it on the
//	 bottom of your library.
//	 Whenever chaos ensues, return target permanent card from your
//	 graveyard to the battlefield."
//
// A plane (ADR 0216 §6). One triggered ability with two conditions. Two
// keyed questions: whether to reveal, then (for a permanent card) whether to
// put it onto the battlefield. A card revealed but not put onto the
// battlefield for any reason (declined, an instant or sorcery, an Aura with
// nothing to enchant, a replacement) goes to the bottom, never staying on
// top (the ruling).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e28deac5-d3c4-48e9-974c-093e3122f317",
		Name:         "The Maelstrom",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenYouPlaneswalkToThisAndAtYourUpkeep("The Maelstrom — you may reveal the top card of your library; a permanent card may be put onto the battlefield, else it goes on the bottom",
				func(g *game.Game, item *game.StackItem) error {
					return MayChoice{
						Question: "The Maelstrom — reveal the top card of your library?",
						Then:     maelstromReveal,
					}.Apply(NewContext(g, item))
				}),
			Targeting(WheneverChaosEnsues("The Maelstrom — return target permanent card from your graveyard to the battlefield",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					id, ok := b16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					return ReturnFromGraveyard{Target: id, Dest: game.ZoneBattlefield, Controller: item.Controller}.Apply(ctx)
				}),
				TargetCardInGraveyard("target permanent card from your graveyard", YouOwn(),
					func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return isPermanentCard(c) })),
		},
	})
}

// maelstromReveal is the first keyed answer. An on-disk key.
var maelstromReveal = MayYes("confirm/the-maelstrom-reveal", func(ctx *Context, r game.ConfirmAnswer) error {
	ids := ctx.Game.RevealTopOfLibraryForEffect(r.Chooser, ctx.Source(), 1, "The Maelstrom — the top card of the library")
	if len(ids) == 0 {
		return nil
	}
	c, ok := ctx.Game.LookupCardForEffect(ids[0])
	if ok && isPermanentCard(c) {
		return MayChoice{
			Player:   r.Chooser,
			Question: "The Maelstrom — put " + c.Name + " onto the battlefield?",
			Then:     maelstromPut,
			Carry:    []uuid.UUID{ids[0]},
		}.Apply(ctx)
	}
	return ctx.Game.PutOnBottomInRandomOrderForEffect(r.Chooser, game.ZoneLibrary, ids)
})

// maelstromPut is the second keyed answer. An on-disk key.
var maelstromPut = MayYesNo("confirm/the-maelstrom-put",
	func(ctx *Context, r game.ConfirmAnswer) error {
		if len(r.Carry) == 0 {
			return nil
		}
		_, _ = ctx.Game.PutFromLibraryOntoBattlefieldForEffect(r.Carry[0], game.LibraryEntryOptions{})
		return maelstromBottomIfLeftOnTop(ctx, r)
	},
	maelstromBottomIfLeftOnTop)

// maelstromBottomIfLeftOnTop puts the revealed card on the bottom unless it
// has left the library (it was put onto the battlefield).
func maelstromBottomIfLeftOnTop(ctx *Context, r game.ConfirmAnswer) error {
	if len(r.Carry) == 0 {
		return nil
	}
	if z := ctx.Game.FindCardZoneForEffect(r.Carry[0]); z == nil || z.Kind != game.ZoneLibrary {
		return nil
	}
	return ctx.Game.PutOnBottomInRandomOrderForEffect(r.Chooser, game.ZoneLibrary, r.Carry[:1])
}
