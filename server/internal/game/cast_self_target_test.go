package game

import (
	"errors"
	"slices"
	"testing"

	"github.com/google/uuid"
)

// cast_self_target_test.go — #4869. A spell cast from a zone is not a
// legal target of its own "target card in <zone>" clauses: CR 601.2a
// moves the card to the stack before CR 601.2c announces its targets,
// so by then it is not in the zone the clause names. Memory's Journey's
// ruling: "If you cast Memory's Journey with flashback, it won't be in
// the graveyard when you choose targets. It can't target itself."

// anyGraveyardCardSpec is "target card in a graveyard".
func anyGraveyardCardSpec() *TargetSpec {
	return &TargetSpec{
		Mode:  "card_in_graveyard",
		Label: "target card in a graveyard",
		Zones: []ZoneKind{ZoneGraveyard},
		Min:   1, Max: 1,
	}
}

// seedSelfTargetingFlashback wires a flashback instant whose one clause is
// "target card in a graveyard", puts it in my graveyard with another card
// beside it, and returns both.
func seedSelfTargetingFlashback(t *testing.T, g *Game, me *Player, oracle string) (spell, other uuid.UUID) {
	t.Helper()
	withCatalogTargetSpec(t, func(o string) *TargetSpec {
		if o == oracle {
			return anyGraveyardCardSpec()
		}
		return nil
	})
	spell = seedFlashbackCard(t, g, me, oracle)
	o := NewCard("Other Card", me.ID)
	o.TypeLine = "Sorcery"
	me.Graveyard.PushTop(o)
	return spell, o.InstanceID
}

func flashbackCastNaming(g *Game, me *Player, spell, target uuid.UUID) error {
	return g.CastSpell(me.ID, spell, CastSpellParams{
		FromZone: "graveyard", AlternativeCost: "flashback",
		Targets: []TargetRef{{Kind: TargetCard, ID: target}},
	})
}

func TestFlashbackSpellCannotTargetItselfInTheGraveyard(t *testing.T) {
	const oracle = "test-4869-self"
	g := newActiveGame(t)
	me := g.Seats[0]
	spell, _ := seedSelfTargetingFlashback(t, g, me, oracle)

	err := flashbackCastNaming(g, me, spell, spell)
	if !errors.Is(err, ErrIllegalTarget) {
		t.Fatalf("flashback cast naming itself: got %v, want ErrIllegalTarget", err)
	}
	if !me.Graveyard.Contains(spell) {
		t.Errorf("the refused cast moved the card out of the graveyard")
	}
}

func TestFlashbackSpellCanTargetAnotherCardInTheSameGraveyard(t *testing.T) {
	const oracle = "test-4869-other"
	g := newActiveGame(t)
	me := g.Seats[0]
	spell, other := seedSelfTargetingFlashback(t, g, me, oracle)

	if err := flashbackCastNaming(g, me, spell, other); err != nil {
		t.Fatalf("flashback cast naming another graveyard card: %v", err)
	}
	if !g.Stack.Contains(spell) {
		t.Errorf("the accepted cast is not on the stack")
	}
}

// The same card cast from hand names a graveyard card as it always did:
// the rule only takes the card being cast out of its own candidates.
func TestHandCastOfTheSameCardIsUnaffected(t *testing.T) {
	const oracle = "test-4869-hand"
	g := newActiveGame(t)
	me := g.Seats[0]
	_, other := seedSelfTargetingFlashback(t, g, me, oracle)
	c := NewCard("Test Looting", me.ID)
	c.TypeLine = "Sorcery"
	c.ManaCost = "{R}"
	c.OracleID = oracle
	me.Hand.PushTop(c)

	if err := g.CastSpell(me.ID, c.InstanceID, CastSpellParams{
		Targets: []TargetRef{{Kind: TargetCard, ID: other}},
	}); err != nil {
		t.Fatalf("hand cast naming a graveyard card: %v", err)
	}
}

// The candidate list the view and the bot build for a cast (SourceCast)
// leaves the card out; the same walk from a source that is not being
// cast (an ability of a card in a graveyard, which stays there while
// it targets) still offers it.
func TestCastCandidateListOmitsTheCardBeingCast(t *testing.T) {
	const oracle = "test-4869-list"
	g := newActiveGame(t)
	me := g.Seats[0]
	spell, other := seedSelfTargetingFlashback(t, g, me, oracle)
	card, ok := g.LookupCardForEffect(spell)
	if !ok {
		t.Fatalf("seeded spell not found")
	}

	var cast, staying LegalTargets
	g.WithWriteLock(func() {
		cast = g.LegalTargetsForEffect(SourceCast(me.ID, &card), anyGraveyardCardSpec())
		staying = g.LegalTargetsForEffect(SourceObject(me.ID, &card), anyGraveyardCardSpec())
	})
	if slices.Contains(cast.Cards, spell) {
		t.Errorf("the cast's candidate list offers the card being cast: %v", cast.Cards)
	}
	if !slices.Contains(cast.Cards, other) {
		t.Errorf("the cast's candidate list lost the other graveyard card: %v", cast.Cards)
	}
	if lt := g.LegalTargetsFor(SourceCast(me.ID, &card), oracle); lt == nil || slices.Contains(lt.Cards, spell) {
		t.Errorf("LegalTargetsFor offers the card being cast: %+v", lt)
	}

	if !slices.Contains(staying.Cards, spell) {
		t.Errorf("a source that is not being cast lost the card from its list: %v", staying.Cards)
	}
}

// A copy of the flashback spell is a new object on the stack (CR 707.10);
// it keeps the original's target and that target is still legal, as
// before #4869.
func TestCopyOfAFlashbackGraveyardSpellKeepsItsTarget(t *testing.T) {
	const oracle = "test-4869-copy"
	g := newActiveGame(t)
	me := g.Seats[0]
	spell, other := seedSelfTargetingFlashback(t, g, me, oracle)
	if err := flashbackCastNaming(g, me, spell, other); err != nil {
		t.Fatalf("flashback cast: %v", err)
	}

	var targets []TargetRef
	var copyErr error
	allIllegal := true
	g.WithWriteLock(func() {
		var copyID uuid.UUID
		copyID, copyErr = g.CopySpellKeepingTargetsForEffect(spell, me.ID)
		if item := g.StackMeta[copyID]; copyErr == nil && item != nil {
			targets = append(targets, item.Targets...)
			allIllegal = spellAllTargetsIllegalLocked(g, item)
		}
	})
	if copyErr != nil {
		t.Fatalf("copy: %v", copyErr)
	}
	if len(targets) != 1 || targets[0].ID != other {
		t.Fatalf("copy targets: got %+v, want the other graveyard card", targets)
	}
	if allIllegal {
		t.Errorf("the copy's kept target reads as illegal")
	}
}
