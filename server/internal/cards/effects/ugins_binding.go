package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ugin's Binding — Instant {2}{U}, devoid:
//
//	"Devoid (This card has no color.)
//	 Return target nonland permanent you don't control to its owner's
//	 hand.
//	 Whenever you cast a colorless spell with mana value 7 or greater,
//	 you may exile this card from your graveyard. When you do, return
//	 each nonland permanent you don't control to its owner's hand."
//
// SANDBOX GAP, DECLARED. Devoid is colour data (the dump's colour list
// is empty), but an empty list also means "not stamped" to the engine,
// which then reads the colours off the mana cost (game.printedColors,
// the gap ADR 0034 names for every devoid card). So this card is blue
// in play, not colourless. Its own clause is unaffected: the spells it
// watches for are other cards, read by their own colours.
//
// The graveyard half is three printed pieces, each its own mechanism:
//
//   - A trigger that watches from the GRAVEYARD (CR 113.6; the card
//     says so), on a colourless spell you cast with mana value 7 or
//     more. A spell with X in its cost counts the announced X on the
//     stack (CR 202.3e).
//   - "You may exile this card" is asked as the trigger RESOLVES
//     (MayChoice, CR 608.2d), and the exile only happens if this card is
//     still in the graveyard as the same object (CR 400.7): one that
//     left and came back is a new card the ability knows nothing
//     about.
//   - "When you do" is a CR 603.12 reflexive trigger, created only by
//     an exile that really happened. It goes on the stack above the
//     parent, so every player has a response window, and returns each
//     nonland permanent the controller does not control, simultaneously.
func init() {
	graveyardTrigger := WheneverYouCast(And(Colorless(), ManaValueGE(7)),
		"Ugin's Binding — you may exile it from your graveyard to return each nonland permanent you don't control",
		uginsBindingMayExile)
	graveyardTrigger.Zones = []game.ZoneKind{game.ZoneGraveyard}

	Register(Spec{
		OracleID:     "ed622e71-f348-4e46-8fb9-05aae430983a",
		Name:         "Ugin's Binding",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"Devoid isn't applied: the game still treats this card as blue."},
		Targets:      TargetPermanent("target nonland permanent you don't control", Nonland(), OpponentControls()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind == game.TargetCard {
					return BounceToHand{Target: t.ID}.Apply(ctx)
				}
			}
			return nil
		},
		Triggered: []game.TriggeredAbility{graveyardTrigger},
	})
}

// uginsBindingBounceAllBody is the reflexive "return each nonland
// permanent you don't control to its owner's hand".
var uginsBindingBounceAllBody = game.SimpleDelayedBody("ugins-binding/bounce-all", uginsBindingBounceAll)

func uginsBindingBounceAll(g *game.Game, item *game.StackItem) error {
	return BounceAllMatching{
		Match: And(Nonland(), OpponentControls()),
	}.Apply(NewContext(g, item))
}

// uginsBindingMayExile asks whether to exile the card, and exiles it.
func uginsBindingMayExile(g *game.Game, item *game.StackItem) error {
	return MayChoice{
		Question: "Ugin's Binding — exile it from your graveyard to return each nonland permanent you don't control to its owner's hand?",
		YesLabel: "Exile it",
		NoLabel:  "Leave it",
		OnYes:    uginsBindingExileThenBounce,
	}.Apply(NewContext(g, item))
}

// uginsBindingExileThenBounce exiles this card from the graveyard, and
// only if that happened creates the reflexive trigger.
func uginsBindingExileThenBounce(ctx *Context) error {
	self := ctx.Source()
	if self == uuid.Nil || !uginsBindingStillInGraveyard(ctx, self) {
		return nil
	}
	if err := ctx.Game.ExileCardForEffect(self); err != nil {
		return err
	}
	return WhenYouDo("Ugin's Binding — return each nonland permanent you don't control to its owner's hand",
		uginsBindingBounceAllBody).Apply(ctx)
}

// uginsBindingStillInGraveyard reports whether `self` sits in a
// graveyard as the object the ability came from: same card, same
// object epoch (CR 400.7).
func uginsBindingStillInGraveyard(ctx *Context, self uuid.UUID) bool {
	z := ctx.Game.FindCardZoneForEffect(self)
	if z == nil || z.Kind != game.ZoneGraveyard {
		return false
	}
	card, ok := ctx.Game.LookupCardForEffect(self)
	if !ok {
		return false
	}
	ref := ctx.Item.SourceObject
	return ref.ID != self || ref.Epoch == card.ObjectEpoch
}
