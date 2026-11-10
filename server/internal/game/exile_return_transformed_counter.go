package game

import "github.com/google/uuid"

// exile_return_transformed_counter.go — "exile it, then put it onto the
// battlefield transformed under its owner's control with a finality counter
// on it" (Esper Origins, Final Fantasy, #4462): ExileAndReturnTransformed's
// shape for a SPELL that moves itself, with an entry counter.
//
// ExileAndReturnTransformedForEffect works on a permanent, the way a Saga's
// chapter III does. A sorcery's own instruction names the spell card on the
// stack: the card is exiled from there (the spell moving itself, CR 608.2n,
// which finishSpellResolutionLocked already reads as "nothing left to
// route"), the back face is set on it in exile (CR 712.8a puts a card in any
// zone but the battlefield and the stack front face up, so this is the one
// moment it is turned), and it returns as a new object (CR 400.7) through
// the ordinary exile-to-battlefield door, so every replacement and entry
// event sees a creature enter. A card whose other face is an instant or a
// sorcery cannot be put onto the battlefield transformed and stays where the
// exile left it (CR 712.14b's shape, as CanTransform asks).

// ExileSpellAndReturnTransformedWithCounterForEffect exiles the card (a spell
// on the stack, or a permanent) and returns it to the battlefield
// transformed under `controller`'s control, uuid.Nil meaning its owner,
// entering with n counters of the named kind (CR 614.1c). A card that is
// gone, or has no permanent back face, is left alone: nothing exiled,
// nothing returned.
//
// Caller must hold g.mu in write mode (resolution frame).
func (g *Game) ExileSpellAndReturnTransformedWithCounterForEffect(cardID, controller uuid.UUID, counter string, n int) error {
	card, ok := g.LookupCardForEffect(cardID)
	if !ok {
		return nil
	}
	front := card
	front.ActiveFace = 0
	if !CanTransform(front) {
		return nil
	}
	return g.ExileCardKeyedForEffect(cardID, KeyedTail{
		Ref: exileReturnTransformedCounter, Player: controller, Carry: []uuid.UUID{cardID},
		N: 1, Word: counter, Nums: []int{n},
	})
}

// exileReturnTransformedCounter is ExileSpellAndReturnTransformedWithCounter's
// second half: once the card reached exile, set its back face and return it
// under Player's control with the counters. A card the exile did not reach
// stays where the replacement put it. An on-disk key: never renamed, never
// reused.
var exileReturnTransformedCounter = RegisterTail("engine/exile-return-transformed-with-counter", func(g *Game, o TailOutcome) error {
	if len(o.Tail.Carry) == 0 || len(o.IDs) != 1 {
		return nil
	}
	id := o.Tail.Carry[0]
	var counters map[string]int
	if o.Tail.Word != "" && len(o.Tail.Nums) > 0 && o.Tail.Nums[0] > 0 {
		counters = map[string]int{o.Tail.Word: o.Tail.Nums[0]}
	}
	setFaceInZoneLocked(g.Exile, id, o.Tail.N)
	_, err := g.returnFromExileToBattlefieldLocked(id, o.Tail.Player, false, counters, entryThen{})
	return err
})
