package game

import (
	"slices"
	"testing"

	"github.com/google/uuid"
)

// cast_tally_card_types_test.go — CastTally.CardTypes (#4677): the card
// types among the spells a player cast this turn (CR 205.2a), read off
// each spell as it became cast (CR 601.2i). April O'Neil, Hacktivist's
// "draw a card for each card type among spells you've cast this turn"
// is the reader; the catalog half is in cards/effects.

// castCardForTest puts c in p's hand, known to p, and casts it with the
// params given, then resolves it, so a later sorcery-speed cast finds an
// empty stack.
func castCardForTest(t *testing.T, g *Game, p *Player, c Card, params CastSpellParams) {
	t.Helper()
	g.WithWriteLock(func() {
		p.Hand.PushTop(c)
		g.markCardKnownInZoneLocked(p.Hand, c.InstanceID)
	})
	if err := g.CastSpell(p.ID, c.InstanceID, params); err != nil {
		t.Fatalf("cast %s: %v", c.Name, err)
	}
	passPriorityUntilResolvedForTest(t, g, c.InstanceID)
}

// typedCard is a plain {0} card with the type line given.
func typedCard(owner uuid.UUID, name, typeLine string) Card {
	c := NewCard(name, owner)
	c.TypeLine = typeLine
	c.ManaCost = "{0}"
	c.Controller = owner
	return c
}

func cardTypesCastFor(g *Game, p *Player) []string {
	var out []string
	g.ReadSnapshot(func() { out = g.CastTallyFor(p.ID).CardTypesCast() })
	return out
}

// The issue's two headline cases: an artifact creature spell is two card
// types (CR 205.2b), and a second creature spell adds none.
func TestCastTallyCountsEachCardTypeAmongSpellsOnce(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	toMainPhase(t, g)

	castCardForTest(t, g, me, typedCard(me.ID, "Fixture Golem", "Artifact Creature — Golem"), CastSpellParams{})
	if got := cardTypesCastFor(g, me); !slices.Equal(got, []string{"Artifact", "Creature"}) {
		t.Fatalf("after an artifact creature spell: %v, want [Artifact Creature]", got)
	}
	castCardForTest(t, g, me, typedCard(me.ID, "Fixture Bear", "Creature — Bear"), CastSpellParams{})
	castCardForTest(t, g, me, typedCard(me.ID, "Fixture Elf", "Legendary Creature — Elf"), CastSpellParams{})
	if got := cardTypesCastFor(g, me); !slices.Equal(got, []string{"Artifact", "Creature"}) {
		t.Fatalf("two more creature spells moved it: %v, want [Artifact Creature]", got)
	}
}

// Kindred is a card type (CR 205.2a, 308.1), so a kindred instant adds two;
// a supertype (CR 205.4a) adds nothing.
func TestCastTallyCountsKindredAsACardType(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	toMainPhase(t, g)

	castCardForTest(t, g, me, typedCard(me.ID, "Fixture Kindred", "Kindred Instant — Elf"), CastSpellParams{})
	castCardForTest(t, g, me, typedCard(me.ID, "Fixture Snow", "Snow Sorcery"), CastSpellParams{})
	if got := cardTypesCastFor(g, me); !slices.Equal(got, []string{"Instant", "Kindred", "Sorcery"}) {
		t.Fatalf("got %v, want [Instant Kindred Sorcery]", got)
	}
}

// A land played is never a spell (CR 305.1): neither a land card nor an
// MDFC played as its land face adds a type, or casts anything.
func TestCastTallyIgnoresALandPlayed(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	toMainPhase(t, g)

	land := sorceryLandMDFC(me.ID)
	g.WithWriteLock(func() {
		me.Hand.PushTop(land)
		g.markCardKnownInZoneLocked(me.Hand, land.InstanceID)
	})
	if err := g.CastSpell(me.ID, land.InstanceID, CastSpellParams{Face: 1}); err != nil {
		t.Fatalf("play the land face: %v", err)
	}
	if findCardForTest(g.Battlefield, land.InstanceID) == nil {
		t.Fatal("the land face was not played")
	}
	var tally CastTally
	g.ReadSnapshot(func() { tally = g.CastTallyFor(me.ID) })
	if tally.Total != 0 || tally.CardTypes != "" {
		t.Fatalf("a land played was counted as a spell: %+v", tally)
	}
}

// The spell's types as it was cast: an MDFC cast as its back face has only
// that face's types on the stack (CR 712.8f), an Adventure only its
// alternative characteristics (CR 715.3b) and a split card only the half
// cast (CR 709.3b).
func TestCastTallyReadsTheFaceThatWasCast(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	toMainPhase(t, g)

	castCardForTest(t, g, me, creatureSorceryMDFC(me.ID), CastSpellParams{Face: 1})
	if got := cardTypesCastFor(g, me); !slices.Equal(got, []string{"Sorcery"}) {
		t.Fatalf("an MDFC cast as its sorcery back: %v, want [Sorcery]", got)
	}
	castCardForTest(t, g, me, adventureFixture(me.ID), CastSpellParams{Face: 1})
	castCardForTest(t, g, me, freeSplit(me.ID), CastSpellParams{Face: 0})
	if got := cardTypesCastFor(g, me); !slices.Equal(got, []string{"Instant", "Sorcery"}) {
		t.Fatalf("after an Adventure and a split half: %v, want [Instant Sorcery], no creature", got)
	}
}

// A copy of a spell isn't cast (CR 707.10), so copying an opponent's spell
// adds nothing to the copier's tally; and an opponent's own spells are in
// their tally, not yours.
func TestCastTallyIgnoresCopiesAndOpponentsSpells(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	toMainPhase(t, g)

	spell := typedCard(opp.ID, "Fixture Bolt", "Instant")
	g.WithWriteLock(func() {
		opp.Hand.PushTop(spell)
		g.markCardKnownInZoneLocked(opp.Hand, spell.InstanceID)
	})
	if err := g.CastSpell(opp.ID, spell.InstanceID, CastSpellParams{}); err != nil {
		t.Fatalf("opponent casts: %v", err)
	}
	g.WithWriteLock(func() {
		if err := g.CopySpellForEffect(spell.InstanceID, me.ID, false, nil); err != nil {
			t.Fatalf("copy: %v", err)
		}
	})
	if top := g.Stack.Cards[len(g.Stack.Cards)-1]; top.InstanceID == spell.InstanceID || !g.StackMeta[top.InstanceID].IsCopy {
		t.Fatal("no copy on top of the stack")
	}
	if got := cardTypesCastFor(g, me); len(got) != 0 {
		t.Fatalf("a copy or an opponent's spell counted for me: %v", got)
	}
	if got := cardTypesCastFor(g, opp); !slices.Equal(got, []string{"Instant"}) {
		t.Fatalf("opponent's tally: %v, want [Instant]", got)
	}
}

// The tally is written in a restore point (additive in snapshot v9) and
// a restore keeps it.
func TestCastTallyCardTypesSurviveASnapshotRoundTrip(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	toMainPhase(t, g)
	castCardForTest(t, g, me, typedCard(me.ID, "Fixture Golem", "Artifact Creature — Golem"), CastSpellParams{})

	_, restored := roundTrip(t, g)
	if got := cardTypesCastFor(restored, me); !slices.Equal(got, []string{"Artifact", "Creature"}) {
		t.Fatalf("restored tally: %v, want [Artifact Creature]", got)
	}
}
