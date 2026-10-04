package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// revealed_hand_discard.go — "Target player reveals their hand. You
// choose a <kind of> card from it. That player discards that card."
// (ADR 0116, #2078). Append-only: the shape's shared vocabulary lives
// here, one card file per card.

// ChooseFromRevealedHand is the whole sentence as one primitive.
// Player reveals their hand to the table, the resolving effect's
// controller chooses Count cards that pass Filter, and Player discards
// them.
//
// Filter is an ordinary CardPredicate — Nonland(), ManaValueGE(3),
// And(Noncreature(), Nonland()) — bound here to the game and the
// chooser, and evaluated against each card as it is in the hand: the
// front face of a double-faced card (CR 712.8a), a split card's
// combined cost (CR 709.4b), X as 0 (CR 202.3e). nil is "a card".
//
// Label is what Filter admits, the way the card prints it ("nonland
// card"); the prompt names it.
//
// When nothing in the hand passes, the hand is still revealed and
// nothing is chosen or discarded (CR 609.3). Apply returns nil either
// way, so the caller's next line — Thoughtseize's "You lose 2 life" —
// always runs.
type ChooseFromRevealedHand struct {
	Player uuid.UUID
	Filter CardPredicate
	Label  string
	// Count defaults to one.
	Count int
}

// Apply queues the reveal and the pick. See the type comment.
func (e ChooseFromRevealedHand) Apply(ctx *Context) error {
	if e.Player == uuid.Nil {
		return nil
	}
	count := e.Count
	if count <= 0 {
		count = 1
	}
	chooser := ctx.Controller()
	var filter func(game.Card) bool
	if e.Filter != nil {
		g, pred := ctx.Game, e.Filter
		filter = func(c game.Card) bool { return pred(g, chooser, c) }
	}
	ctx.Game.QueueDiscardFromRevealedHand(game.RevealedHandDiscard{
		Chooser:    chooser,
		FromPlayer: e.Player,
		Source:     ctx.Source(),
		Count:      count,
		Filter:     filter,
		Label:      e.Label,
	})
	return nil
}

// TargetedPlayer is the player the spell's first target clause chose,
// if that target is still legal (CR 608.2b), or uuid.Nil.
func TargetedPlayer(ctx *Context) uuid.UUID {
	t, ok := ctx.ClauseTarget(0)
	if !ok || t.Kind != game.TargetPlayer {
		return uuid.Nil
	}
	return t.ID
}

// TargetRevealsYouChooseDiscard is the shape's most common card body:
// the targeted player reveals, you choose a card that passes filter,
// they discard it — and nothing else. Duress, Distress, Unmask,
// Pelakka Predation.
func TargetRevealsYouChooseDiscard(filter CardPredicate, label string) func(*game.StackItem, *Context) error {
	return func(_ *game.StackItem, ctx *Context) error {
		return ChooseFromRevealedHand{Player: TargetedPlayer(ctx), Filter: filter, Label: label}.Apply(ctx)
	}
}

// ModeTargetRevealsYouChooseDiscard is TargetRevealsYouChooseDiscard as
// a modal bullet's body: the player is this mode occurrence's own
// target, re-checked at resolution (CR 608.2b). Mardu Charm, Auntie's
// Sentence, Poison the Waters and the pool's other charms (ADR 0116
// pool B).
func ModeTargetRevealsYouChooseDiscard(filter CardPredicate, label string) func(*game.StackItem, *Context, int) error {
	return func(_ *game.StackItem, ctx *Context, occ int) error {
		t, ok := ModeTarget(ctx, occ)
		if !ok || t.Kind != game.TargetPlayer {
			return nil
		}
		return ChooseFromRevealedHand{Player: t.ID, Filter: filter, Label: label}.Apply(ctx)
	}
}

// NonlandPermanentCard is the pick's "nonland permanent card": an
// artifact, creature, enchantment, planeswalker or battle card, read as
// it is in the hand. Auntie's Sentence, Dai Li Indoctrination.
func NonlandPermanentCard() CardPredicate { return And(Nonland(), Permanent()) }
