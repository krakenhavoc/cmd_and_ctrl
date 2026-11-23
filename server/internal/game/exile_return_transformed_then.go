package game

import "github.com/google/uuid"

// exile_return_transformed_then.go — "exile it, then return it to the
// battlefield transformed under its owner's control. If you do, …" (Liliana,
// Heretical Healer, #5948): ExileAndReturnTransformedForEffect with the rest
// of the effect as a registered tail, run only once the card has really come
// back, told its NEW instance ID as TailOutcome.IDs (CR 400.7).

// ExileAndReturnTransformedThenForEffect exiles the permanent and returns it
// to the battlefield transformed under controller's control (uuid.Nil: its
// owner's), then runs `then`. `then` is told the new object's ID; it does not
// run when the exile or the return did not happen (a replacement redirected
// either, or the permanent is gone or cannot transform). A question the entry
// asks is a restore point: nothing here holds a closure.
//
// Caller must hold g.mu in write mode (resolution frame).
func (g *Game) ExileAndReturnTransformedThenForEffect(cardID, controller uuid.UUID, then KeyedTail) error {
	card := findBattlefieldCard(g, cardID)
	if card == nil || !CanTransform(*card) {
		return nil
	}
	face := 1 - card.ActiveFace
	return g.ExileCardKeyedForEffect(cardID, KeyedTail{
		Ref: exileReturnTransformedThen, Player: controller, Carry: []uuid.UUID{cardID}, N: face, Next: &then,
	})
}

// exileReturnTransformedThen is the second half: once the card reached
// exile, set its other face and return it with the card's own tail waiting on
// the entry. An on-disk key: never renamed, never reused.
var exileReturnTransformedThen = RegisterTail("engine/exile-return-transformed-then", func(g *Game, o TailOutcome) error {
	if len(o.Tail.Carry) == 0 || len(o.IDs) != 1 || o.Tail.Next == nil {
		return nil
	}
	id := o.Tail.Carry[0]
	setFaceInZoneLocked(g.Exile, id, o.Tail.N)
	_, err := g.returnFromExileToBattlefieldLocked(id, o.Tail.Player, false, nil, entryThen{keyed: *o.Tail.Next})
	return err
})
