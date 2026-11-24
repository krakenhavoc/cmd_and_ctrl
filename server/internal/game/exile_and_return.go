package game

import "github.com/google/uuid"

// exile_and_return.go — "exile that card instead of putting it into your
// graveyard as it resolves. If you do, return it to your hand at the
// beginning of the next end step" (Feather, the Redeemed).
//
// Like exile_and_plot.go it is the spell's own resolution exit
// (resolution_exits.go): it never fires for a countered or fizzled spell,
// and it competes with the other exits under CR 616.1. The spell is marked
// while it is on the stack; once the card is really in exile a delayed
// trigger (CR 603.7) is created for the next end step, about that object
// (CR 400.7): a card that left exile before then is not returned.

// MarkSpellExilesAndReturnsForEffect marks the spell `spellID` so that it is
// exiled, and returned to its controller's hand at the next end step, instead
// of going to its owner's graveyard as it resolves. Reports false when the ID
// is not a spell still on the stack, or is a copy (a copy is not a card,
// CR 707.10a, and ceases to exist instead).
//
// Caller must hold g.mu.
func (g *Game) MarkSpellExilesAndReturnsForEffect(spellID uuid.UUID) bool {
	_, item, ok := g.stolenSpellLocked(spellID)
	if !ok || item.IsCopy {
		return false
	}
	item.ExileAndReturn = true
	return true
}

// exiledAndReturnTail schedules the return once the card is really in exile.
// The returner is the spell's controller. An on-disk key: never renamed,
// never reused.
var exiledAndReturnTail = RegisterTail("engine/exiled-and-return", func(g *Game, o TailOutcome) error {
	if len(o.Tail.Carry) == 0 {
		return nil
	}
	c := exiledCardByIDLocked(g, o.Tail.Carry[0])
	if c == nil {
		return nil
	}
	g.ScheduleDelayedTriggerForEffect(DelayedTrigger{
		Controller:   o.Tail.Player,
		SourceCardID: c.InstanceID,
		Label:        "Return " + c.Name + " to its owner's hand",
		At:           StepEnd,
		Body:         exileReturnBody,
		Params:       EffectParams{Object: ObjectRef{ID: c.InstanceID, Epoch: c.ObjectEpoch}},
	})
	return nil
})

// exileReturnBody is the delayed trigger's resolution: the card goes to its
// owner's hand if it is still the object that was exiled. Registered in init,
// like reboundCastBody, because it reaches the move machinery. An on-disk key.
var exileReturnBody BodyRef

func init() {
	exileReturnBody = DelayedBody("exile-and-return/hand", func(g *Game, _ *StackItem, p EffectParams) error {
		c := exiledCardByIDLocked(g, p.Object.ID)
		if c == nil || c.ObjectEpoch != p.Object.Epoch {
			return nil
		}
		return g.BounceToHandForEffect(p.Object.ID)
	})
}
