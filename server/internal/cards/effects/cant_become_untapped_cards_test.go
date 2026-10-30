package effects

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cant_become_untapped_cards_test.go — #4574: "can't become untapped"
// (CR 614.17b), proven on Frozen in Ice and Spider-Woman, Secret Agent. Act
// of Heroism ("Untap target creature") is the untap effect that must do
// nothing to a locked creature.

const (
	oracleFrozenInIce = "82fa14d8-5019-4e1b-a86c-923d9292b87a"
	oracleSpiderWoman = "04e53f36-9f16-4678-9caf-9afda85f76b4"
)

func TestCantBecomeUntappedCardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oracleFrozenInIce: "Frozen in Ice",
		oracleSpiderWoman: "Spider-Woman, Secret Agent",
	} {
		spec, ok := Lookup(oracle)
		if !ok || spec.Name != name {
			t.Fatalf("Lookup(%s) = %q, %v", oracle, spec.Name, ok)
		}
		if spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: completeness %s, caveats %v", name, spec.Completeness, spec.Caveats)
		}
	}
	spec, _ := Lookup(oracleSpiderWoman)
	if !slices.Contains(spec.PrintedKeywords, "flash") {
		t.Errorf("Spider-Woman keywords %v, want flash", spec.PrintedKeywords)
	}
}

// untapSpell casts Act of Heroism at id and lets it resolve.
func untapSpell(t *testing.T, g *game.Game, id uuid.UUID) {
	t.Helper()
	castCatalogSpell(t, g, "Act of Heroism", "Instant", actOfHeroismOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: id}})
	passPriorityAroundTable(t, g)
}

func TestFrozenInIceTapsLocksAndLosesAbilitiesUntilItLeaves(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	host := pushPermanentForTest(g, me.ID, "Host", "", "Creature — Bear")
	aura := castCatalogSpell(t, g, "Frozen in Ice", auraTypeLine, oracleFrozenInIce,
		[]game.TargetRef{{Kind: game.TargetCard, ID: host}})
	passPriorityAroundTable(t, g)
	if got := attachmentHostOf(t, g, aura); got.Kind != game.TargetCard || got.ID != host {
		t.Fatalf("Frozen in Ice attached to %+v, want host %s", got, host)
	}
	if !tappedForTest(t, g, host) {
		t.Fatal("Frozen in Ice's entry trigger did not tap the enchanted creature")
	}

	// The untap step leaves it tapped.
	advanceToUpkeepOf(t, g, 0)
	if !tappedForTest(t, g, host) {
		t.Fatal("the enchanted creature untapped during its controller's untap step")
	}
	// An untap effect does nothing to it (the point of "can't become
	// untapped", against Claustrophobia's "doesn't untap").
	untapSpell(t, g, host)
	if !tappedForTest(t, g, host) {
		t.Fatal("Act of Heroism untapped a creature that can't become untapped")
	}
	// Nor does the sandbox untap.
	if err := g.TapCard(host, false); err != nil {
		t.Fatal(err)
	}
	if !tappedForTest(t, g, host) {
		t.Fatal("the sandbox untap untapped a creature that can't become untapped")
	}

	// The wire says so, so the client offers no untap.
	var never bool
	g.WithWriteLock(func() {
		c := battlefieldCardPtr(t, g, host)
		never = g.CantBecomeUntappedCheckerLocked()(c)
	})
	if !never {
		t.Error("the lock is not readable for the view")
	}

	// Once the Aura leaves the creature untaps normally.
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, aura); err != nil {
		t.Fatal(err)
	}
	untapSpell(t, g, host)
	if tappedForTest(t, g, host) {
		t.Fatal("the creature stayed tapped after Frozen in Ice left")
	}
	if err := g.TapCard(host, true); err != nil {
		t.Fatal(err)
	}
	advanceToUpkeepOf(t, g, 0)
	if tappedForTest(t, g, host) {
		t.Fatal("the creature missed its untap step after Frozen in Ice left")
	}
}

func TestSpiderWomanTapsAndLocksUntilYouLoseHer(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	opp := g.Seats[1]
	target := pushCreatureToBattlefieldForTest(g, opp.ID, "Target")
	mine := pushCreatureToBattlefieldForTest(g, me.ID, "Mine")
	spider := castHoldCreature(t, g, "Spider-Woman, Secret Agent",
		"Legendary Creature — Spider Human Spy Hero", oracleSpiderWoman, 1, 4)
	if opts := pickTargetPrompt(t, g).PickTargetCards; slices.Contains(opts, mine) || slices.Contains(opts, spider) || !slices.Contains(opts, target) {
		t.Fatalf("Spider-Woman's target options = %v, want only the opponent's creature", opts)
	}
	answerPickTarget(t, g, target)
	passPriorityAroundTable(t, g)
	if !tappedForTest(t, g, target) {
		t.Fatal("Spider-Woman's trigger did not tap its target")
	}

	for round := 0; round < 2; round++ {
		advanceToUpkeepOf(t, g, 1)
		if !tappedForTest(t, g, target) {
			t.Fatalf("round %d: the locked creature untapped in its controller's untap step", round+1)
		}
		advanceToUpkeepOf(t, g, 0)
	}
	untapSpell(t, g, target)
	if !tappedForTest(t, g, target) {
		t.Fatal("Act of Heroism untapped a creature that can't become untapped")
	}

	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, spider); err != nil {
		t.Fatal(err)
	}
	untapSpell(t, g, target)
	if tappedForTest(t, g, target) {
		t.Fatal("an untap effect did nothing after Spider-Woman left the battlefield")
	}
}

// Leaving in response to the trigger taps the creature with no lock: the
// duration never starts (CR 611.2b).
func TestSpiderWomanLeavingInResponseTapsWithoutALock(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	target := pushCreatureToBattlefieldForTest(g, opp.ID, "Target")
	spider := castHoldCreature(t, g, "Spider-Woman, Secret Agent",
		"Legendary Creature — Spider Human Spy Hero", oracleSpiderWoman, 1, 4)
	answerPickTarget(t, g, target)
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, spider); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	if !tappedForTest(t, g, target) {
		t.Fatal("the trigger did not tap its target")
	}
	if holdCount(t, g, target) != 0 {
		t.Fatal("a lock was recorded although its duration never started")
	}
	untapSpell(t, g, target)
	if tappedForTest(t, g, target) {
		t.Fatal("the creature could not untap although no lock was ever recorded")
	}
}
