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

	// The variants (#2115, ADR 0116's 2026-10-05 amendment). See
	// game.RevealedHandDiscard for each; leaving all four zero is the
	// plain discard above.

	// Exile exiles the chosen card instead of discarding it: "You
	// choose a … card from it and exile that card" (Appetite for
	// Brains). Not a discard, so no madness and no discard triggers.
	Exile bool
	// Optional is "You may choose".
	Optional bool
	// FromGraveyard adds "or a card from their graveyard" (Agonizing
	// Remorse). Filter reads the hand only.
	FromGraveyard bool
	// Then is the rest of the card, told what was chosen: a
	// RevealedPickThen or RevealedPickFirst var declared in the card's
	// file.
	Then game.RevealedPickThen
	// Measure reads a number off each candidate while the spell is
	// still resolving (Talara's Bane's toughness), for Then to read
	// back off the chosen card as pick.Measures. See
	// game.RevealedHandDiscard.Measure for why it is read here.
	Measure func(g *game.Game, c game.Card) int
}

// Apply queues the reveal and the pick. See the type comment.
//
// With a Then and nothing to choose, Then runs before Apply returns,
// and its error is Apply's.
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
	d := game.RevealedHandDiscard{
		Chooser:       chooser,
		FromPlayer:    e.Player,
		Source:        ctx.Source(),
		Count:         count,
		Filter:        filter,
		Label:         e.Label,
		Optional:      e.Optional,
		FromGraveyard: e.FromGraveyard,
		Then:          e.Then,
	}
	if e.Exile {
		d.Destination = game.PickExile
	}
	if e.Measure != nil {
		g, measure := ctx.Game, e.Measure
		d.Measure = func(c game.Card) int { return measure(g, c) }
	}
	if e.Exile || e.Optional || e.FromGraveyard || e.Then.Key() != "" || e.Measure != nil {
		_, err := ctx.Game.RevealedHandPickForEffect(d)
		return err
	}
	ctx.Game.QueueDiscardFromRevealedHand(d)
	return nil
}

// RevealedPickThen registers a revealed-hand pick's continuation that
// runs once the chosen card has moved, as a card file's package-level
// var. The body is handed a Context rebuilt from values — the chooser
// as its controller, the card that asked as its source — never from
// the stack item it began with, which has long resolved; and the pick.
// The key is an on-disk identity ("revealed-pick/<card>-<what>"): never
// renamed, never reused.
func RevealedPickThen(key string, body func(ctx *Context, pick game.RevealedPick) error) game.RevealedPickThen {
	return game.RegisterRevealedPickThen(key, func(g *game.Game, r game.RevealedPick) error {
		return body(revealedPickContext(g, r), r)
	})
}

// RevealedPickFirst is RevealedPickThen for a clause printed BEFORE the
// move, which reads the chosen card while it is still in the hand
// (Talara's Bane). The body must call pick.Done exactly once.
func RevealedPickFirst(key string, body func(ctx *Context, pick game.RevealedPick) error) game.RevealedPickThen {
	return game.RegisterRevealedPickFirst(key, func(g *game.Game, r game.RevealedPick) error {
		return body(revealedPickContext(g, r), r)
	})
}

// revealedPickContext is the resolving spell's context as a pick's
// continuation sees it, rebuilt from values (invertPolarityContext's
// shape).
func revealedPickContext(g *game.Game, r game.RevealedPick) *Context {
	return NewContext(g, &game.StackItem{
		Kind:         game.StackItemSpell,
		Controller:   r.Chooser,
		Owner:        r.Chooser,
		SourceCardID: r.Source,
	})
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

// TargetRevealsYouChooseDiscardAbility is TargetRevealsYouChooseDiscard
// as a triggered or activated ability's Effect: Grief's enters trigger,
// Mind Slash's sacrifice ability, a Saga's chapter. "You" is the
// ability's controller (CR 113.8): the player who activated it, or the
// player who controlled the source when it triggered. That is
// ctx.Controller(), the stack item's controller, so the chooser is the
// same whether the source is still on the battlefield or a cost
// sacrificed it (Pilfering Imp). The targeted player is the ability's
// first target clause.
func TargetRevealsYouChooseDiscardAbility(filter CardPredicate, label string) Effect {
	body := TargetRevealsYouChooseDiscard(filter, label)
	return func(g *game.Game, item *game.StackItem) error {
		return body(item, NewContext(g, item))
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
