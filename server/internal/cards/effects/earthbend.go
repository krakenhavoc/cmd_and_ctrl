package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// earthbend.go — the catalog side of the AVATAR keyword action
// earthbend (#1178). The verb itself is one engine function
// (`game.EarthbendForEffect`, server/internal/game/earthbend.go); this
// file is the two lines a card needs to print it.
//
//	Earthbend N. (Target land you control becomes a 0/0 creature with
//	haste that's still a land. Put N +1/+1 counters on it. When it
//	dies or is exiled, return it to the battlefield tapped.)
//
// Twelve cards print it and all twelve print the same reminder text,
// which is the whole argument for a keyword action rather than a
// composition per card: the animation, the haste, the counters and the
// delayed return are ONE verb, and a card is the count plus whatever
// else its own sentence says. `Airbend` is the same shape one keyword
// earlier.
//
// EVERY EARTHBEND CARD SHARES ONE TARGET CLAUSE, so it is written
// once here. "Target land you control" is part of the keyword, not of
// the card: a card that wants a different clause is not printing
// earthbend.

// EarthbendTargets is the target clause every earthbend card prints:
// "target land you control" (CR 115.4 — the target is chosen at
// announce and re-checked at resolution, so a land that stops being a
// land or changes hands in response takes the spell with it).
//
// Declared as a constructor rather than a package var because
// `game.TargetSpec` is a pointer and `Register` keeps the one it is
// given: two cards sharing a value would share whatever a later
// mutation did to it.
func EarthbendTargets() *game.TargetSpec {
	return TargetPermanent("target land you control", And(Land(), YouControl()))
}

// Earthbend is "Earthbend N" applied to one named land.
//
// N is the number of +1/+1 counters, and it is read at RESOLUTION for
// every card that computes it ("earthbend X, where X is the number of
// Forests you control") — the caller does the counting and hands over
// an int, exactly as `DrawCards.N` works.
//
// Zero is legal and is NOT a no-op: the land still becomes a 0/0
// creature with haste and still gets the delayed return, so it dies to
// the toughness state-based action and comes back tapped. That is what
// Rockalanche does with no Forests on the battlefield, and the engine
// keeps the distinction (`KeywordAction.actsAtZeroCount`).
type Earthbend struct {
	// Target is the land. A zero value earthbends nothing, which is
	// how "up to one target" and a fizzled slot are spelled.
	Target uuid.UUID

	// N is the number of +1/+1 counters to put on it.
	N int

	// Then is the rest of the sentence after the verb — Earthshape's
	// "Then each creature you control with power less than or equal to
	// that land's power gains hexproof". It runs once the counters have
	// LANDED (#1282): on a board with two different counter
	// replacements that is after the CR 616 ordering prompt is
	// answered, not when Apply returns, which is why it is a field and
	// not the next statement. It is handed a fresh Context over the
	// live game, on the continuation contract every other *Then
	// primitive follows. Nil means nothing follows.
	Then func(ctx *Context) error
}

func (e Earthbend) Apply(ctx *Context) error {
	if e.Target == uuid.Nil || ctx.isNewSourceObject(e.Target) { // #1432
		return nil
	}
	var then func(g *game.Game) error
	if e.Then != nil {
		rest, item := e.Then, ctx.Item
		then = func(g *game.Game) error { return rest(NewContext(g, item)) }
	}
	return ctx.Game.EarthbendThenForEffect(ctx.Controller(), ctx.Source(), e.Target, e.N, then)
}

// EarthbendFirstTarget is the whole body of a card whose only sentence
// is "Earthbend N": earthbend the first still-legal card target on the
// resolving item.
//
// `count` is handed the resolution context so a card can compute its
// own N — Rockalanche counts Forests, The Legend of Kyoshi's chapter
// II counts cards in hand, and Earthbending Lesson ignores it and
// returns 4. Nil means zero, which is the earthbend-0 case above and
// not an error.
//
// A card with a second clause about the same land (Kyoshi's "that land
// becomes an Island", Earthshape's hexproof sweep) calls `Earthbend`
// directly and keeps its own target read, because it needs the land's
// ID for the second clause too.
func EarthbendFirstTarget(count func(ctx *Context) int) func(*game.StackItem, *Context) error {
	return func(_ *game.StackItem, ctx *Context) error {
		target := FirstLegalBattlefieldTarget(ctx)
		if target == uuid.Nil {
			return nil
		}
		n := 0
		if count != nil {
			n = count(ctx)
		}
		return Earthbend{Target: target, N: n}.Apply(ctx)
	}
}

// EarthbendCount is the fixed-N spelling: `EarthbendFirstTarget(
// EarthbendCount(4))` is the whole of Earthbending Lesson.
func EarthbendCount(n int) func(*Context) int {
	return func(*Context) int { return n }
}

// FirstLegalBattlefieldTarget is the first card target on the
// resolving item that is still legal (CR 608.2b) and still on the
// battlefield.
//
// The two checks are not the same one. `LegalTargets` re-runs the
// announced clause, which is what CR 608.2b asks for; the battlefield
// test is what a primitive that is about to mutate a permanent needs,
// because a clause with no structured spec falls back to existence
// alone and a card in a graveyard exists. Airbending Lesson makes the
// same pair inline; this is that idiom with a name.
func FirstLegalBattlefieldTarget(ctx *Context) uuid.UUID {
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard || t.ID == uuid.Nil {
			continue
		}
		if onBattlefield(ctx.Game, t.ID) {
			return t.ID
		}
	}
	return uuid.Nil
}
