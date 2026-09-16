package game

import (
	"testing"

	"github.com/google/uuid"
)

// escape_test.go — S29, the escape half of "cast from somewhere other
// than hand".
//
// The cast PATH is pinned by cast_zones_test.go and is shared with
// flashback; what is new here is the PRICE. Escape is the first cost
// in the engine with a multi-card component ("exile five other cards
// from your graveyard"), and every test below is about a way that
// count can be cheated:
//
//   - naming too few, or too many
//   - naming the same card five times
//   - naming the spell itself, which the printed word "other"
//     excludes and which CR 601.2a excludes anyway
//   - naming cards out of somebody else's graveyard
//
// Every one of those fails in the player's favour if it is not
// checked, which is the direction a sandbox must never err in. The
// last two tests are about CR 702.138c's counters, and the contrast
// with flashback that trips people up: an escaped card is NOT exiled
// when it leaves the stack.

// escapeCost is the constructor's shape, spelled out here so the
// game package can test the mechanic without importing the effects
// package — the same trick flashbackOffer plays.
func escapeCost(cost string, n int) AlternativeCost {
	return AlternativeCost{
		Key:      "escape",
		Label:    "Escape—" + cost,
		ManaCost: cost,
		FromZone: ZoneGraveyard,
		ExileFromGraveyard: &TargetSpec{
			Mode:  "card_in_graveyard",
			Label: "other cards from your graveyard",
			Zones: []ZoneKind{ZoneGraveyard},
			CardOK: func(_ *Game, caster uuid.UUID, c Card, _ ZoneKind) bool {
				return c.Owner == caster
			},
			Min: n, Max: n,
		},
	}
}

// fodder pushes n plain cards into a player's graveyard and returns
// their IDs, newest last.
func fodder(p *Player, n int) []uuid.UUID {
	out := make([]uuid.UUID, 0, n)
	for i := 0; i < n; i++ {
		c := NewCard("Fodder", p.ID)
		c.TypeLine = "Instant"
		p.Graveyard.PushTop(c)
		out = append(out, c.InstanceID)
	}
	return out
}

// seedEscapeSpell wires an escape sorcery with an `exile` cost of n
// and puts one in the active seat's graveyard, along with enough
// fodder to pay for it twice over.
func seedEscapeSpell(t *testing.T, g *Game, me *Player, oracle string, n int) (uuid.UUID, []uuid.UUID) {
	t.Helper()
	withCatalogCastableZones(t, castableZonesFor(oracle, ZoneGraveyard))
	withCatalogAlternativeCosts(t, altCostFor(oracle, escapeCost("{1}{B}", n)))
	id := looterInGraveyard(t, g, me, oracle)
	return id, fodder(me, n+2)
}

func castEscape(g *Game, me *Player, id uuid.UUID, pay []uuid.UUID) error {
	return g.CastSpell(me.ID, id, CastSpellParams{
		FromZone:        "graveyard",
		AlternativeCost: "escape",
		AltCostIDs:      pay,
	})
}

// The happy path. The spell reaches the stack and exactly the named
// cards are in exile — the cost is paid at announce, not at
// resolution, which is what makes it a cost.
func TestEscapeCastExilesTheNamedCards(t *testing.T) {
	const oracle = "test-escape"
	g := newActiveGame(t)
	me := g.Seats[0]
	id, yard := seedEscapeSpell(t, g, me, oracle, 3)
	pay := yard[:3]

	if err := castEscape(g, me, id, pay); err != nil {
		t.Fatalf("escape cast: %v", err)
	}
	if !g.Stack.Contains(id) {
		t.Fatalf("escape cast did not reach the stack")
	}
	for _, p := range pay {
		if !g.Exile.Contains(p) {
			t.Errorf("a named card was not exiled")
		}
		if me.Graveyard.Contains(p) {
			t.Errorf("a named card is still in the graveyard")
		}
	}
	for _, left := range yard[3:] {
		if !me.Graveyard.Contains(left) {
			t.Errorf("an unnamed card left the graveyard — the cost took more than it asked for")
		}
	}
	if item := g.StackMeta[id]; item == nil || item.AltCost != "escape" {
		t.Errorf("the stack item does not remember the escape cost")
	}
}

// Too few. The interesting half is the second assertion: a rejected
// cast must not leave a half-paid cost behind.
func TestEscapeRejectsTooFewCards(t *testing.T) {
	const oracle = "test-escape"
	g := newActiveGame(t)
	me := g.Seats[0]
	id, yard := seedEscapeSpell(t, g, me, oracle, 3)

	if err := castEscape(g, me, id, yard[:2]); err == nil {
		t.Fatalf("a two-card payment bought a three-card escape cost")
	}
	if !me.Graveyard.Contains(id) {
		t.Errorf("the rejected cast moved the spell out of the graveyard")
	}
	for _, p := range yard[:2] {
		if !me.Graveyard.Contains(p) {
			t.Errorf("the rejected cast exiled a card anyway")
		}
	}
}

// Too many. Overpaying is rejected rather than trimmed: a client
// that sent four for a three-card cost is confused about the card,
// and quietly eating the extra is a worse answer than an error.
func TestEscapeRejectsTooManyCards(t *testing.T) {
	const oracle = "test-escape"
	g := newActiveGame(t)
	me := g.Seats[0]
	id, yard := seedEscapeSpell(t, g, me, oracle, 3)

	if err := castEscape(g, me, id, yard[:4]); err == nil {
		t.Fatalf("a four-card payment was accepted for a three-card cost")
	}
	if g.Exile.Size() != 0 {
		t.Errorf("the rejected cast exiled something")
	}
}

// The same card five times is the cheapest imaginable escape, and
// the one a naive "len(ids) == n" check lets through.
func TestEscapeRejectsTheSameCardTwice(t *testing.T) {
	const oracle = "test-escape"
	g := newActiveGame(t)
	me := g.Seats[0]
	id, yard := seedEscapeSpell(t, g, me, oracle, 3)

	if err := castEscape(g, me, id, []uuid.UUID{yard[0], yard[0], yard[1]}); err == nil {
		t.Fatalf("a duplicate payment paid the escape cost")
	}
	if !me.Graveyard.Contains(yard[0]) {
		t.Errorf("the rejected cast exiled a card anyway")
	}
}

// "OTHER cards." The spell is on its way to the stack and is not one
// of them — CR 601.2a is the rule, and the word on the card is the
// same fact stated from the player's side.
func TestEscapeCannotPayWithItself(t *testing.T) {
	const oracle = "test-escape"
	g := newActiveGame(t)
	me := g.Seats[0]
	id, yard := seedEscapeSpell(t, g, me, oracle, 3)

	if err := castEscape(g, me, id, []uuid.UUID{id, yard[0], yard[1]}); err == nil {
		t.Fatalf("the escaping card paid part of its own escape cost")
	}
	if g.Exile.Contains(id) {
		t.Errorf("the rejected cast exiled the spell itself")
	}
}

// "YOUR graveyard." Escape does not eat the table's graveyards, and
// the zone lookup is what enforces it — the cost is matched against
// the caster's own graveyard rather than against every graveyard the
// spec's zone list would otherwise reach.
func TestEscapeCannotPayFromAnotherPlayersGraveyard(t *testing.T) {
	const oracle = "test-escape"
	g := newActiveGame(t)
	me, them := g.Seats[0], g.Seats[1]
	id, yard := seedEscapeSpell(t, g, me, oracle, 3)
	theirs := fodder(them, 1)

	if err := castEscape(g, me, id, []uuid.UUID{yard[0], yard[1], theirs[0]}); err == nil {
		t.Fatalf("an escape cost was paid out of an opponent's graveyard")
	}
	if !them.Graveyard.Contains(theirs[0]) {
		t.Errorf("the rejected cast exiled an opponent's card")
	}
}

// seedEscapeCreature is seedEscapeSpell for a permanent, with the
// CR 702.138c counter clause attached to the cost.
func seedEscapeCreature(t *testing.T, g *Game, me *Player, oracle string, counters int) (uuid.UUID, []uuid.UUID) {
	t.Helper()
	cost := escapeCost("{1}{G}", 2)
	if counters > 0 {
		cost.EntersWithCounterName = "+1/+1"
		cost.EntersWithCounterCount = counters
	}
	withCatalogCastableZones(t, castableZonesFor(oracle, ZoneGraveyard))
	withCatalogAlternativeCosts(t, altCostFor(oracle, cost))
	advanceTo(t, g, StepPrecombatMain)
	c := NewCard("Test Typhon", me.ID)
	c.TypeLine = "Creature — Snake Beast"
	c.ManaCost = "{2}{G}{G}"
	c.Power, c.Toughness = 4, 4
	c.OracleID = oracle
	me.Graveyard.PushTop(c)
	return c.InstanceID, fodder(me, 2)
}

// CR 702.138c: "this creature escapes with three +1/+1 counters on
// it". The counters ride the entry, so they are on the permanent
// before anything else looks at it.
func TestEscapedCreatureEntersWithItsCounters(t *testing.T) {
	const oracle = "test-typhon"
	g := newActiveGame(t)
	me := g.Seats[0]
	id, yard := seedEscapeCreature(t, g, me, oracle, 3)

	if err := castEscape(g, me, id, yard); err != nil {
		t.Fatalf("escape cast: %v", err)
	}
	resolveTop(t, g)
	c, ok := g.LookupCardForEffect(id)
	if !ok || !g.Battlefield.Contains(id) {
		t.Fatalf("the escaped creature did not reach the battlefield")
	}
	if got := c.Counters["+1/+1"]; got != 3 {
		t.Errorf("+1/+1 counters: got %d, want 3", got)
	}
}

// The same creature cast for its printed cost out of hand enters
// with none. The clause hangs off the COST, which is the whole
// reason it is declared on the AlternativeCost rather than on the
// card.
func TestHardCastCreatureEntersWithoutEscapeCounters(t *testing.T) {
	const oracle = "test-typhon"
	g := newActiveGame(t)
	me := g.Seats[0]
	id, _ := seedEscapeCreature(t, g, me, oracle, 3)
	// Move it out of the graveyard and into hand, then cast it
	// normally.
	moved, err := MoveCard(me.Graveyard, me.Hand, id)
	if err != nil {
		t.Fatalf("seed to hand: %v", err)
	}
	if err := g.CastSpell(me.ID, moved.InstanceID, CastSpellParams{}); err != nil {
		t.Fatalf("hand cast: %v", err)
	}
	resolveTop(t, g)
	c, ok := g.LookupCardForEffect(id)
	if !ok {
		t.Fatalf("the creature vanished")
	}
	if got := c.Counters["+1/+1"]; got != 0 {
		t.Errorf("a hard-cast creature entered with %d escape counters", got)
	}
}

// The contrast with flashback, and the one people get wrong: escape
// has no CR 702.34a clause, so an escaped spell goes to the
// graveyard like any other and can escape again for another helping
// of cards.
func TestEscapedSpellReturnsToTheGraveyard(t *testing.T) {
	const oracle = "test-escape"
	g := newActiveGame(t)
	me := g.Seats[0]
	id, yard := seedEscapeSpell(t, g, me, oracle, 3)

	if err := castEscape(g, me, id, yard[:3]); err != nil {
		t.Fatalf("escape cast: %v", err)
	}
	resolveTop(t, g)
	if g.Exile.Contains(id) {
		t.Fatalf("an escaped spell was exiled — that is flashback's clause, not escape's")
	}
	if !me.Graveyard.Contains(id) {
		t.Errorf("an escaped spell did not return to the graveyard")
	}
}
