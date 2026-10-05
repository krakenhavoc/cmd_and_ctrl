package effects

import "github.com/google/uuid"

// revealTopReadingManaValue is the first half of "reveal the top card
// of your library and put that card into your hand", with the mana
// value the card's follow-up reads — Yuriko's drain, Stronghold
// Arena's "you lose life equal to its mana value". It reveals (so no
// EventDrawCard: a draw payoff does not see it) and reads the mana
// value, and leaves the move to the caller, because whether a clause
// is gated on the card reaching the hand is the card's, not the
// shape's: Yuriko's drain is unconditional, Stronghold Arena's "if you
// do" is not (BounceToHand.Then).
//
// The mana value is read before the move, for Dark Confidant's reason:
// the last moment the card is certainly findable where the effect found
// it (CR 202.3e makes the number the same in either zone). An empty
// library reveals nothing and returns uuid.Nil.
func revealTopReadingManaValue(ctx *Context, reason string) (card uuid.UUID, manaValue int, err error) {
	var revealed []uuid.UUID
	if err := (RevealTopOfLibrary{
		Player:   ctx.Controller(),
		N:        1,
		Reason:   reason,
		Revealed: &revealed,
	}).Apply(ctx); err != nil {
		return uuid.Nil, 0, err
	}
	if len(revealed) == 0 {
		return uuid.Nil, 0, nil
	}
	if c, ok := ctx.Game.LookupCardForEffect(revealed[0]); ok {
		manaValue = c.ManaValue()
	}
	return revealed[0], manaValue, nil
}
