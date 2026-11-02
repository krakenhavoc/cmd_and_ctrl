package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// msc_last_a_cards_test.go — Marvel Super Heroes Commander, last A batch
// (#4455): Klaw, Master of Sound; Elektra, Femme Fatale; Loyal Guardian.
// Each test plays the real catalog entry and asserts what its text says.

const (
	mscLAKlawOracle    = "8eaa8833-ab9e-4bd8-92de-63e195b8d912"
	mscLAElektraOracle = "75f2545d-cd31-4404-953f-1110f646228b"
	mscLAGuardianOrcl  = "de1cfa83-2049-48ab-b093-2f120a3ef4f2"
)

func TestMscLAEveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		mscLAKlawOracle:    "Klaw, Master of Sound",
		mscLAElektraOracle: "Elektra, Femme Fatale",
		mscLAGuardianOrcl:  "Loyal Guardian",
	}
	for oracle, name := range want {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// --- Klaw, Master of Sound -----------------------------------------------

const mscLAKlawType = "Legendary Creature — Elemental Rogue Villain"

func TestMscLAKlawHasDeathtouch(t *testing.T) {
	g := newCatalogGame(t)
	klaw := b12Push(g, g.Seats[0].ID, "Klaw, Master of Sound", mscLAKlawType, mscLAKlawOracle, 3, 3)
	assertKeywords(t, g, klaw, "deathtouch")
}

func TestMscLAKlawExilesTheTopCardFaceDownForYouToPlayWithAnyMana(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	klaw := b27Push(g, me.ID, "Klaw, Master of Sound", mscLAKlawType, mscLAKlawOracle, "{2}{B}", 3, 3)
	top := opp.Library.Cards[len(opp.Library.Cards)-1].InstanceID

	attackWith(t, g, opp.ID, klaw)
	passPriorityAroundTable(t, g)

	if z := e2Zone(g, top); z != game.ZoneExile {
		t.Fatalf("the damaged player's top card is in %q, want exile", z)
	}
	perm := e2Permission(g, top)
	if perm == nil || perm.Player != me.ID || perm.CastOnly || !perm.AnyType {
		t.Errorf("permission %+v, want an any-type play (not cast-only) grant to Klaw's controller", perm)
	}
	e2AssertFaceDownFor(t, g, top, me.ID)
}

func TestMscLAKlawDoesNothingForCombatDamageToAnotherCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b27Push(g, me.ID, "Klaw, Master of Sound", mscLAKlawType, mscLAKlawOracle, "{2}{B}", 3, 3)
	before := len(opp.Library.Cards)
	// Noncombat damage to a player is not combat damage.
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventDealDamage, Source: uuid.New(), Target: opp.ID, Amount: 3, Actor: me.ID})
	})
	passPriorityAroundTable(t, g)
	if len(opp.Library.Cards) != before {
		t.Errorf("library %d -> %d: only Klaw's combat damage exiles a card", before, len(opp.Library.Cards))
	}
}

func TestMscLAKlawGainsIndestructibleWhenYouPlayACardFromExile(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	klaw := b12Push(g, me.ID, "Klaw, Master of Sound", mscLAKlawType, mscLAKlawOracle, 3, 3)

	// A spell cast from hand does not.
	castCatalogSpell(t, g, "Bolt", "Instant", "", nil)
	passPriorityAroundTable(t, g)
	if hasEffectiveKeyword(t, g, klaw, "indestructible") {
		t.Fatal("a cast from hand gave indestructible")
	}

	shock := b20HandCard(me, "Shock", "Instant")
	g.WithWriteLock(func() {
		_ = g.ExileCardWithPermissionForEffect(shock, game.CastPermission{Player: me.ID})
	})
	if err := g.CastSpell(me.ID, shock, game.CastSpellParams{FromZone: "exile"}); err != nil {
		t.Fatalf("casting from exile: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !hasEffectiveKeyword(t, g, klaw, "indestructible") {
		t.Error("playing a card from exile gives Klaw indestructible")
	}
}

// --- Elektra, Femme Fatale -----------------------------------------------

const mscLAElektraType = "Legendary Creature — Human Assassin Hero"

func TestMscLAElektraDealsTwoToYouThenFourToTheCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := b12Creature(g, opp.ID, "Big Bear", "Creature — Bear", 5, 5)
	life := me.Life

	castHoldCreature(t, g, "Elektra, Femme Fatale", mscLAElektraType, mscLAElektraOracle, 3, 2)
	assistCardSettle(t, g, me.ID, &game.TargetRef{Kind: game.TargetCard, ID: victim}, true)

	if me.Life != life-2 {
		t.Errorf("life %d, want %d: Elektra deals 2 damage to you", me.Life, life-2)
	}
	if c := battlefieldCardPtr(t, g, victim); c.DamageMarked != 4 {
		t.Errorf("damage on the creature %d, want 4", c.DamageMarked)
	}
}

func TestMscLAElektraDecliningDealsNoDamageAtAll(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := b12Creature(g, opp.ID, "Big Bear", "Creature — Bear", 5, 5)
	life := me.Life

	castHoldCreature(t, g, "Elektra, Femme Fatale", mscLAElektraType, mscLAElektraOracle, 3, 2)
	assistCardSettle(t, g, me.ID, &game.TargetRef{Kind: game.TargetCard, ID: victim}, false)

	if me.Life != life {
		t.Errorf("life %d, want %d: declining deals no damage to you", me.Life, life)
	}
	if c := battlefieldCardPtr(t, g, victim); c.DamageMarked != 0 {
		t.Errorf("damage on the creature %d, want 0: the reflexive trigger needs the 2 damage", c.DamageMarked)
	}
}

// --- Loyal Guardian ------------------------------------------------------

const mscLAGuardianType = "Creature — Rhino"

func TestMscLALoyalGuardianHasTrample(t *testing.T) {
	g := newCatalogGame(t)
	guardian := b12Push(g, g.Seats[0].ID, "Loyal Guardian", mscLAGuardianType, mscLAGuardianOrcl, 4, 4)
	assertKeywords(t, g, guardian, "trample")
}

func TestMscLALoyalGuardianWithoutYourCommanderPutsNoCounters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	guardian := b12Push(g, me.ID, "Loyal Guardian", mscLAGuardianType, mscLAGuardianOrcl, 4, 4)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	advanceTo(t, g, game.StepBeginCombat)
	passPriorityAroundTable(t, g)
	if plusOnes(g, guardian) != 0 || plusOnes(g, bear) != 0 {
		t.Errorf("counters %d / %d, want none without your commander", plusOnes(g, guardian), plusOnes(g, bear))
	}
}

func TestMscLALoyalGuardianWithYourCommanderPutsACounterOnEachOfYourCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	guardian := b12Push(g, me.ID, "Loyal Guardian", mscLAGuardianType, mscLAGuardianOrcl, 4, 4)
	commander := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "My Commander",
		TypeLine: "Legendary Creature — Bear", Power: 2, Toughness: 5, Owner: me.ID, Controller: me.ID, IsCommander: true})
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	advanceTo(t, g, game.StepBeginCombat)
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{guardian, commander} {
		if got := plusOnes(g, id); got != 1 {
			t.Errorf("creature %s has %d counters, want 1", id, got)
		}
	}
	if got := plusOnes(g, theirs); got != 0 {
		t.Errorf("an opponent's creature got %d counters, want 0", got)
	}
}

func TestMscLALoyalGuardianChecksTheCommanderAgainOnResolution(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	guardian := b12Push(g, me.ID, "Loyal Guardian", mscLAGuardianType, mscLAGuardianOrcl, 4, 4)
	commander := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "My Commander",
		TypeLine: "Legendary Creature — Bear", Power: 2, Toughness: 5, Owner: me.ID, Controller: me.ID, IsCommander: true})
	advanceTo(t, g, game.StepBeginCombat)
	// The trigger is on the stack; the commander leaves before it resolves
	// (CR 603.4: the condition is checked again on resolution).
	g.WithWriteLock(func() { _ = g.ExileCardForEffect(commander) })
	passPriorityAroundTable(t, g)
	if got := plusOnes(g, guardian); got != 0 {
		t.Errorf("Loyal Guardian has %d counters, want 0: the commander left before the trigger resolved", got)
	}
}
