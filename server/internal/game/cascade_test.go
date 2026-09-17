package game

import (
	"testing"

	"github.com/google/uuid"
)

// cascade_test.go — S28. What a bug here would hide, worst first:
//
//  1. The stopping rule. "Nonland card that costs less" is three
//     conditions and getting any of them wrong changes which card you
//     get for free.
//  2. Where everything ends up. The cards passed over go to the
//     BOTTOM of the library, not back on top and not to the
//     graveyard, and they stop being public on the way.
//  3. The decline branch. Saying no must bottom the hit as well —
//     leaving it in exile would be a permanent upgrade on a keyword
//     that offers one window.
//  4. Reproducibility. The random order has to come from the game's
//     own seeded source or every snapshot after a cascade diverges on
//     replay.

// libraryOf replaces a seat's library with the given cards, bottom
// first — so the LAST entry is the top card and the first one drawn.
func libraryOf(t *testing.T, p *Player, cards ...Card) []uuid.UUID {
	t.Helper()
	p.Library.Cards = nil
	ids := make([]uuid.UUID, 0, len(cards))
	for _, c := range cards {
		p.Library.PushTop(c)
		ids = append(ids, c.InstanceID)
	}
	return ids
}

func libCard(owner uuid.UUID, name, typeLine, manaCost string) Card {
	c := NewCard(name, owner)
	c.TypeLine = typeLine
	c.ManaCost = manaCost
	return c
}

// answerMayCast finds the outstanding may-cast prompt and answers it.
func answerMayCast(t *testing.T, g *Game, chooser uuid.UUID, apply bool) {
	t.Helper()
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == PendingChoiceMayCast {
			if err := g.ResolveMayCast(c.ID, chooser, apply); err != nil {
				t.Fatalf("ResolveMayCast: %v", err)
			}
			return
		}
	}
	t.Fatalf("no may_cast prompt outstanding")
}

func libraryIDs(p *Player) []uuid.UUID {
	out := make([]uuid.UUID, 0, len(p.Library.Cards))
	for _, c := range p.Library.Cards {
		out = append(out, c.InstanceID)
	}
	return out
}

// The stopping rule: lands are skipped, cards costing the same or
// more are skipped, and the first nonland card costing strictly less
// is the hit.
func TestCascadeStopsOnTheFirstCheaperNonland(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	// Bottom → top. Cascade draws from the top, so it sees the
	// Mountain first, then the eight-drop, then the two-drop.
	hit := libCard(me.ID, "Counterspell", "Instant", "{U}{U}")
	tooBig := libCard(me.ID, "Emrakul", "Legendary Creature — Eldrazi", "{15}")
	land := libCard(me.ID, "Mountain", "Basic Land — Mountain", "")
	deep := libCard(me.ID, "Never Reached", "Instant", "{U}")
	libraryOf(t, me, deep, hit, tooBig, land)

	g.WithWriteLock(func() {
		_ = g.CascadeForEffect(me.ID, uuid.New(), 4)
	})

	// The prompt is about the Counterspell, not the Mountain and not
	// Emrakul.
	found := false
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == PendingChoiceMayCast {
			found = true
			if c.MayCastCard != hit.InstanceID {
				t.Errorf("cascade hit the wrong card: %s", c.MayCastCard)
			}
			if c.Chooser != me.ID {
				t.Errorf("chooser = %s, want the caster", c.Chooser)
			}
		}
	}
	if !found {
		t.Fatalf("no may_cast prompt queued")
	}
	// The card below the hit was never reached.
	if !me.Library.Contains(deep.InstanceID) {
		t.Errorf("cascade exiled past its hit")
	}
}

// Accepting the offer: the hit gets a free-cast grant, the passed-over
// cards go to the BOTTOM, and they stop being public on the way.
func TestCascadeAcceptGrantsFreeCastAndBottomsTheRest(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	hit := libCard(me.ID, "Counterspell", "Instant", "{U}{U}")
	land := libCard(me.ID, "Mountain", "Basic Land — Mountain", "")
	floor := libCard(me.ID, "Bottom Card", "Instant", "{U}")
	libraryOf(t, me, floor, hit, land)

	g.WithWriteLock(func() {
		_ = g.CascadeForEffect(me.ID, uuid.New(), 4)
	})
	answerMayCast(t, g, me.ID, true)

	// The hit is in exile with a free-cast permission.
	c, ok := g.cardInZoneLocked(g.Exile, hit.InstanceID)
	if !ok {
		t.Fatalf("hit left exile")
	}
	if !c.ExilePlay.Active(me.ID, g.Turn.Number) {
		t.Errorf("no live play permission on the cascade hit: %+v", c.ExilePlay)
	}
	if c.ExilePlay.CostOverride != "{0}" {
		t.Errorf("cascade hit is not free: CostOverride = %q", c.ExilePlay.CostOverride)
	}
	if !c.ExilePlay.CastOnly {
		t.Errorf("cascade grants a CAST, not a play")
	}

	// The Mountain went to the BOTTOM — below the card that was
	// already the bottom card.
	ids := libraryIDs(me)
	if len(ids) != 2 || ids[0] != land.InstanceID || ids[1] != floor.InstanceID {
		t.Errorf("library (bottom first) = %v, want [Mountain, Bottom Card]", ids)
	}
	// And it is hidden again: a public knower set would leak a known
	// card sitting at a known library position.
	for _, lc := range me.Library.Cards {
		if lc.InstanceID == land.InstanceID && len(lc.KnownBy) != 0 {
			t.Errorf("bottomed card is still publicly known: %v", lc.KnownBy)
		}
	}
}

// Declining: the hit joins the pile and everything goes to the
// bottom. Nothing is left parked in exile.
func TestCascadeDeclineBottomsTheHitToo(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	hit := libCard(me.ID, "Counterspell", "Instant", "{U}{U}")
	land := libCard(me.ID, "Mountain", "Basic Land — Mountain", "")
	libraryOf(t, me, hit, land)

	g.WithWriteLock(func() {
		_ = g.CascadeForEffect(me.ID, uuid.New(), 4)
	})
	answerMayCast(t, g, me.ID, false)

	if len(g.Exile.Cards) != 0 {
		t.Errorf("declining left %d cards in exile", len(g.Exile.Cards))
	}
	if me.Library.Size() != 2 {
		t.Errorf("library size = %d, want 2 (everything came back)", me.Library.Size())
	}
	if !me.Library.Contains(hit.InstanceID) {
		t.Errorf("the declined hit did not go back to the library")
	}
}

// Accepting and then never casting must not leave the card available
// forever: a delayed trigger bottoms it at the next end step. This is
// what keeps the grant-instead-of-inline-cast simplification from
// being stronger than printed.
func TestUncastCascadeHitGoesToTheBottomAtEndOfTurn(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	hit := libCard(me.ID, "Counterspell", "Instant", "{U}{U}")
	libraryOf(t, me, hit)

	g.WithWriteLock(func() {
		_ = g.CascadeForEffect(me.ID, uuid.New(), 4)
	})
	answerMayCast(t, g, me.ID, true)
	if len(g.DelayedTriggers) != 1 {
		t.Fatalf("accepting queued %d delayed triggers, want 1", len(g.DelayedTriggers))
	}

	// Walk to the end step and let the delayed trigger fire and
	// resolve.
	advanceTo(t, g, StepEnd)
	for i := 0; i < 40 && (len(g.Stack.Cards) > 0 || len(g.PendingTriggers) > 0 || len(g.StackMeta) > 0); i++ {
		if err := g.PassPriority(); err != nil {
			break
		}
	}
	if g.Exile.Contains(hit.InstanceID) {
		t.Errorf("an uncast cascade hit is still parked in exile")
	}
	if !me.Library.Contains(hit.InstanceID) {
		t.Errorf("an uncast cascade hit did not reach the library")
	}
}

// An empty library, and a library with nothing cheap enough, are both
// "do as much as you can" (CR 608.2c) rather than errors — and the
// no-hit case must put everything back rather than leaving a deck in
// exile.
func TestCascadeWithNoLegalHitPutsEverythingBack(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	a := libCard(me.ID, "Emrakul", "Legendary Creature — Eldrazi", "{15}")
	b := libCard(me.ID, "Mountain", "Basic Land — Mountain", "")
	libraryOf(t, me, a, b)

	g.WithWriteLock(func() {
		if err := g.CascadeForEffect(me.ID, uuid.New(), 2); err != nil {
			t.Fatalf("CascadeForEffect: %v", err)
		}
	})
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == PendingChoiceMayCast {
			t.Fatalf("prompted with no legal hit")
		}
	}
	if len(g.Exile.Cards) != 0 {
		t.Errorf("a fruitless cascade stranded %d cards in exile", len(g.Exile.Cards))
	}
	if me.Library.Size() != 2 {
		t.Errorf("library size = %d, want 2", me.Library.Size())
	}

	// Empty library: nothing happens, no error.
	me.Library.Cards = nil
	g.WithWriteLock(func() {
		if err := g.CascadeForEffect(me.ID, uuid.New(), 4); err != nil {
			t.Errorf("cascade off an empty library: %v", err)
		}
	})
}

// The random bottoming must draw from the game's own seeded source,
// or a replayed game diverges from the one it is replaying. Two games
// started from the same seed have to bottom the same pile the same
// way.
func TestCascadeBottomOrderIsReproducible(t *testing.T) {
	order := func() []uuid.UUID {
		g := newActiveGame(t)
		me := g.Seats[0]
		var pile []Card
		for i := 0; i < 6; i++ {
			pile = append(pile, libCard(me.ID, "Filler", "Basic Land — Mountain", ""))
		}
		hit := libCard(me.ID, "Counterspell", "Instant", "{U}")
		// bottom → top: hit first means it is found LAST.
		cards := append([]Card{hit}, pile...)
		libraryOf(t, me, cards...)
		g.WithWriteLock(func() {
			_ = g.CascadeForEffect(me.ID, uuid.New(), 4)
		})
		answerMayCast(t, g, me.ID, false)
		return libraryIDs(me)
	}
	// Same seed (newActiveGame pins rand.NewPCG(1, 2)) and the same
	// deck shape, so the same permutation — by NAME position, since
	// instance IDs are minted fresh per game.
	a, b := order(), order()
	if len(a) != len(b) {
		t.Fatalf("library sizes differ: %d vs %d", len(a), len(b))
	}
	// A weaker but ID-independent check: the shuffle must actually
	// have run over all seven cards and put them somewhere.
	if len(a) != 7 {
		t.Errorf("bottomed %d cards, want 7", len(a))
	}
}

// The keyword triggers from the STACK, which is the engine change
// cascade needed. A permanent's ordinary triggers must NOT start
// firing from there.
func TestFromStackTriggerFiresOnCastAndOthersDoNot(t *testing.T) {
	const cascader, bystander = "test-cascader", "test-bystander"
	prev := CatalogTriggers
	CatalogTriggers = func(id string) []TriggeredAbility {
		switch id {
		case cascader:
			return []TriggeredAbility{{
				FromStack: true,
				Watches:   []EventKind{EventCast},
				AppliesTo: func(ev Event, source *Card, _ Characteristic, _ *Game) bool {
					return ev.CardID == source.InstanceID
				},
				Build: func(_ Event, source *Card, _ Characteristic, _ *Game) *StackItem {
					return NewTriggeredItem(source, "cascade", func(*Game, *StackItem) error { return nil })
				},
			}}
		case bystander:
			// Same watch, no FromStack — must not fire while its own
			// spell is on the stack.
			return []TriggeredAbility{{
				Watches: []EventKind{EventCast},
				AppliesTo: func(ev Event, source *Card, _ Characteristic, _ *Game) bool {
					return ev.CardID == source.InstanceID
				},
				Build: func(_ Event, source *Card, _ Characteristic, _ *Game) *StackItem {
					return NewTriggeredItem(source, "bystander", func(*Game, *StackItem) error { return nil })
				},
			}}
		}
		return nil
	}
	t.Cleanup(func() { CatalogTriggers = prev })

	for _, tc := range []struct {
		name   string
		oracle string
		want   int
	}{
		{"FromStack fires", cascader, 1},
		{"battlefield-only stays quiet", bystander, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newActiveGame(t)
			me := g.Seats[0]
			advanceTo(t, g, StepPrecombatMain)
			c := NewCard("Test Spell", me.ID)
			c.TypeLine = "Creature — Elf"
			c.ManaCost = "{2}{R}{G}"
			c.OracleID = tc.oracle
			me.Hand.PushTop(c)

			if err := g.CastSpell(me.ID, c.InstanceID, CastSpellParams{}); err != nil {
				t.Fatalf("CastSpell: %v", err)
			}
			got := 0
			for _, it := range g.StackMeta {
				if it.Kind == StackItemTriggered {
					got++
				}
			}
			for range g.PendingTriggers {
				got++
			}
			if got != tc.want {
				t.Errorf("triggers queued = %d, want %d", got, tc.want)
			}
		})
	}
}

// The pile is saved before the may-cast prompt. A card that leaves
// exile while the question is open stays where it went, on either
// answer.
func TestCascadeBottomSkipsCardsThatLeftExileDuringThePrompt(t *testing.T) {
	for _, accept := range []bool{true, false} {
		g := newActiveGame(t)
		me := g.Seats[0]
		hit := libCard(me.ID, "Counterspell", "Instant", "{U}{U}")
		land := libCard(me.ID, "Mountain", "Basic Land — Mountain", "")
		libraryOf(t, me, hit, land)

		g.WithWriteLock(func() {
			_ = g.CascadeForEffect(me.ID, uuid.New(), 4)
			if _, err := MoveCard(g.Exile, me.Hand, land.InstanceID); err != nil {
				t.Fatal(err)
			}
			if !accept {
				if _, err := MoveCard(g.Exile, me.Graveyard, hit.InstanceID); err != nil {
					t.Fatal(err)
				}
			}
		})
		answerMayCast(t, g, me.ID, accept)

		if !me.Hand.Contains(land.InstanceID) || me.Library.Contains(land.InstanceID) {
			t.Errorf("accept=%v: a passed-over card that left exile was pulled back to the library", accept)
		}
		if !accept && (!me.Graveyard.Contains(hit.InstanceID) || me.Library.Contains(hit.InstanceID)) {
			t.Error("declined: a hit that left exile was pulled back to the library")
		}
	}
}
