package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Expel from Orazca — Instant {1}{U} (EDHREC rank 10782):
//
//	"Ascend (If you control ten or more permanents, you get the city's
//	 blessing for the rest of the game.)
//	 Return target nonland permanent to its owner's hand. If you have
//	 the city's blessing, you may put that permanent on top of its
//	 owner's library instead."
//
// Without the blessing it is a plain bounce. With it the caster is
// asked, and "yes" tucks the permanent on top of its OWNER's library
// (a token ceases to exist either way). Declining returns it to hand,
// which is the printed alternative. The blessing is read as the spell
// resolves, after its own ascend check (CR 702.131a), so ten permanents
// reached by the spell's own resolution already count.
//
// The permanent is the spell's target and is re-checked on the answer: a
// target that left the battlefield in the meantime is left alone. A
// commander headed for a hand or a library is offered the command zone
// by the move (CR 903.9a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "a6e12be3-4166-4bd7-8254-e28b7c8588c7",
		Name:            "Expel from Orazca",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordAscend},
		Targets:         TargetPermanent("target nonland permanent", Nonland()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok || t.Kind != game.TargetCard || !ctx.IsTargetLegal(t) {
				return nil
			}
			if !YouHaveTheCitysBlessing(ctx.Game, item.Controller) {
				return BounceToHand{Target: t.ID}.Apply(ctx)
			}
			return MayChoice{
				Player:   item.Controller,
				Question: "Expel from Orazca — put that permanent on top of its owner's library instead of returning it to hand?",
				YesLabel: "Top of library",
				NoLabel:  "Return to hand",
				OnYes:    expelToTopOfLibrary(t.ID),
				OnNo:     expelToHand(t.ID),
			}.Apply(ctx)
		},
	})
}

// expelToTopOfLibrary is the "yes" branch. It closes over one scalar,
// the StackItem.Effect contract, so an undo across the prompt resolves
// it against the restored game.
func expelToTopOfLibrary(id uuid.UUID) func(ctx *Context) error {
	return func(ctx *Context) error {
		if z := ctx.Game.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneBattlefield {
			return nil
		}
		return ctx.Game.TuckToLibraryForEffect(id, false)
	}
}

// expelToHand is the "no" branch: the printed bounce.
func expelToHand(id uuid.UUID) func(ctx *Context) error {
	return func(ctx *Context) error {
		if z := ctx.Game.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneBattlefield {
			return nil
		}
		return BounceToHand{Target: id}.Apply(ctx)
	}
}
