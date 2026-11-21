package testsbr

import (
	"slices"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestCantBecomeUntappedCardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		effectstest.OracleFrozenInIce: "Frozen in Ice",
		effectstest.OracleSpiderWoman: "Spider-Woman, Secret Agent",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok || spec.Name != name {
			t.Fatalf("Lookup(%s) = %q, %v", oracle, spec.Name, ok)
		}
		if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: completeness %s, caveats %v", name, spec.Completeness, spec.Caveats)
		}
	}
	spec, _ := effects.Lookup(effectstest.OracleSpiderWoman)
	if !slices.Contains(spec.PrintedKeywords, "flash") {
		t.Errorf("Spider-Woman keywords %v, want flash", spec.PrintedKeywords)
	}
}

func TestFrozenInIceTapsLocksAndLosesAbilitiesUntilItLeaves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	host := effectstest.PushPermanentForTest(g, me.ID, "Host", "", "Creature — Bear")
	aura := effectstest.CastCatalogSpell(t, g, "Frozen in Ice", effectstest.AuraTypeLine, effectstest.OracleFrozenInIce,
		[]game.TargetRef{{Kind: game.TargetCard, ID: host}})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.AttachmentHostOf(t, g, aura); got.Kind != game.TargetCard || got.ID != host {
		t.Fatalf("Frozen in Ice attached to %+v, want host %s", got, host)
	}
	if !effectstest.TappedForTest(t, g, host) {
		t.Fatal("Frozen in Ice's entry trigger did not tap the enchanted creature")
	}

	// The untap step leaves it tapped.
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	if !effectstest.TappedForTest(t, g, host) {
		t.Fatal("the enchanted creature untapped during its controller's untap step")
	}
	// An untap effect does nothing to it (the point of "can't become
	// untapped", against Claustrophobia's "doesn't untap").
	effectstest.UntapSpell(t, g, host)
	if !effectstest.TappedForTest(t, g, host) {
		t.Fatal("Act of Heroism untapped a creature that can't become untapped")
	}
	// Nor does the sandbox untap.
	if err := g.TapCard(host, false); err != nil {
		t.Fatal(err)
	}
	if !effectstest.TappedForTest(t, g, host) {
		t.Fatal("the sandbox untap untapped a creature that can't become untapped")
	}

	// The wire says so, so the client offers no untap.
	var never bool
	g.WithWriteLock(func() {
		c := effectstest.BattlefieldCardPtr(t, g, host)
		never = g.CantBecomeUntappedCheckerLocked()(c)
	})
	if !never {
		t.Error("the lock is not readable for the view")
	}

	// Once the Aura leaves the creature untaps normally.
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, aura); err != nil {
		t.Fatal(err)
	}
	effectstest.UntapSpell(t, g, host)
	if effectstest.TappedForTest(t, g, host) {
		t.Fatal("the creature stayed tapped after Frozen in Ice left")
	}
	if err := g.TapCard(host, true); err != nil {
		t.Fatal(err)
	}
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	if effectstest.TappedForTest(t, g, host) {
		t.Fatal("the creature missed its untap step after Frozen in Ice left")
	}
}

func TestSpiderWomanTapsAndLocksUntilYouLoseHer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	opp := g.Seats[1]
	target := effectstest.PushCreatureToBattlefieldForTest(g, opp.ID, "Target")
	mine := effectstest.PushCreatureToBattlefieldForTest(g, me.ID, "Mine")
	spider := effectstest.CastHoldCreature(t, g, "Spider-Woman, Secret Agent",
		"Legendary Creature — Spider Human Spy Hero", effectstest.OracleSpiderWoman, 1, 4)
	if opts := effectstest.PickTargetPrompt(t, g).PickTargetCards; slices.Contains(opts, mine) || slices.Contains(opts, spider) || !slices.Contains(opts, target) {
		t.Fatalf("Spider-Woman's target options = %v, want only the opponent's creature", opts)
	}
	effectstest.AnswerPickTarget(t, g, target)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.TappedForTest(t, g, target) {
		t.Fatal("Spider-Woman's trigger did not tap its target")
	}

	for round := 0; round < 2; round++ {
		effectstest.AdvanceToUpkeepOf(t, g, 1)
		if !effectstest.TappedForTest(t, g, target) {
			t.Fatalf("round %d: the locked creature untapped in its controller's untap step", round+1)
		}
		effectstest.AdvanceToUpkeepOf(t, g, 0)
	}
	effectstest.UntapSpell(t, g, target)
	if !effectstest.TappedForTest(t, g, target) {
		t.Fatal("Act of Heroism untapped a creature that can't become untapped")
	}

	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, spider); err != nil {
		t.Fatal(err)
	}
	effectstest.UntapSpell(t, g, target)
	if effectstest.TappedForTest(t, g, target) {
		t.Fatal("an untap effect did nothing after Spider-Woman left the battlefield")
	}
}

// Leaving in response to the trigger taps the creature with no lock: the
// duration never starts (CR 611.2b).
func TestSpiderWomanLeavingInResponseTapsWithoutALock(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	target := effectstest.PushCreatureToBattlefieldForTest(g, opp.ID, "Target")
	spider := effectstest.CastHoldCreature(t, g, "Spider-Woman, Secret Agent",
		"Legendary Creature — Spider Human Spy Hero", effectstest.OracleSpiderWoman, 1, 4)
	effectstest.AnswerPickTarget(t, g, target)
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, spider); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.TappedForTest(t, g, target) {
		t.Fatal("the trigger did not tap its target")
	}
	if effectstest.HoldCount(t, g, target) != 0 {
		t.Fatal("a lock was recorded although its duration never started")
	}
	effectstest.UntapSpell(t, g, target)
	if effectstest.TappedForTest(t, g, target) {
		t.Fatal("the creature could not untap although no lock was ever recorded")
	}
}
