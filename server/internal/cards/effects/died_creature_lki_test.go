package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// died_creature_lki_test.go — #1675, CR 603.10a: "whenever a creature
// dies" is judged on what the permanent was as it last existed on the
// battlefield. A crewed Vehicle and an animated land are an artifact
// and a land again in the graveyard, so diedCreature reads the types
// the exit stamped on the EventLTB (Event.LastKnownTypes), never the
// graveyard card.

// animateLandForTest makes a land a 3/3 creature that is still a land
// until end of turn — the manland shape (Mutavault), through the same
// scoped-effect record a card's resolution would register.
func animateLandForTest(t *testing.T, g *game.Game, land uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() {
		if !g.RegisterScopedEffectForEffect(uuid.Nil, g.PinnedObjectsLocked(land),
			[]game.Mod{game.AddTypesMod("Creature"), game.SetBasePowerMod(3), game.SetBaseToughnessMod(3)},
			g.UntilEndOfTurnDuration(), "test — the land becomes a 3/3 creature") {
			t.Fatal("setup: the animation registered nothing")
		}
		g.RecomputeLayersIfStaleLocked()
	})
	if c, ok := battlefieldCardByID(g, land); !ok || !c.IsCreature() || !c.IsLand() {
		t.Fatal("setup: the animated land is not a land creature")
	}
}

// The issue's case: a crewed Vehicle attacks, is blocked by something
// bigger, and dies to combat damage. It was a creature when it died,
// so Blood Artist triggers — though the card in the graveyard is an
// artifact with no creature type.
func TestBloodArtistTriggersWhenACrewedVehicleDiesInCombat(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
	pushCatalogPermanent(g, me.ID, "Blood Artist", "Creature — Vampire", bloodArtistOracle, false)
	caravan := pushVehicleForTest(g, me.ID, "Cultivator's Caravan", cultivatorsCaravanOrcl, 5, 5)
	crewForTest(t, g, me.ID, caravan, pushCrewerForTest(g, me.ID, "Crewer", 3))
	wall := b12Creature(g, opp.ID, "Their Wall", "Creature — Wall", 6, 6)
	meBefore, oppBefore := me.Life, opp.Life

	declareAttack(t, g, opp.ID, caravan)
	b35Block(t, g, wall, caravan)
	advanceTo(t, g, game.StepCombatDamage)

	if passUntilPickTarget(t, g, me.ID) == nil {
		t.Fatal("Blood Artist did not trigger on the crewed Vehicle's death")
	}
	if !me.Graveyard.Contains(caravan) {
		t.Fatal("setup: the Caravan should have died to the Wall")
	}
	if c, _ := g.LookupCardForEffect(caravan); c.IsCreature() {
		t.Fatal("setup: the Vehicle in the graveyard should be an artifact only — the point of the test")
	}
	pickPlayer(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)
	if opp.Life != oppBefore-1 || me.Life != meBefore+1 {
		t.Errorf("life: me %d → %d, opp %d → %d; want +1 / -1", meBefore, me.Life, oppBefore, opp.Life)
	}
}

// An uncrewed Vehicle is an artifact on the battlefield too: its death
// is not a creature's.
func TestBloodArtistIgnoresAnUncrewedVehicleDestroyed(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Blood Artist", "Creature — Vampire", bloodArtistOracle, false)
	caravan := pushVehicleForTest(g, me.ID, "Cultivator's Caravan", cultivatorsCaravanOrcl, 5, 5)
	lifeBefore := lifeSnapshot(g)

	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(caravan) })
	if !me.Graveyard.Contains(caravan) {
		t.Fatal("setup: the Caravan was not destroyed")
	}
	if passUntilPickTarget(t, g, me.ID) != nil {
		t.Fatal("Blood Artist triggered on an uncrewed Vehicle")
	}
	passPriorityAroundTable(t, g)
	for i, p := range g.Seats {
		if p.Life != lifeBefore[i] {
			t.Errorf("seat %d life %d → %d; an artifact dying drains nobody", i, lifeBefore[i], p.Life)
		}
	}
}

// An animated land that dies was a creature: Zulaport Cutthroat's
// "whenever a creature you control dies" drains.
func TestZulaportCountsAnAnimatedLandDying(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Zulaport Cutthroat", "Creature — Human Rogue Ally", zulaportCutthroatOracle, false)
	land := b12Permanent(g, me.ID, "Test Manland", "Land")
	animateLandForTest(t, g, land)
	lifeBefore := lifeSnapshot(g)

	b18Kill(t, g, land)
	if c, _ := g.LookupCardForEffect(land); c.IsCreature() {
		t.Fatal("setup: the land in the graveyard should not be a creature")
	}
	for i, p := range g.Seats {
		want := lifeBefore[i] - 1
		if i == 0 {
			want = lifeBefore[i] + 1
		}
		if p.Life != want {
			t.Errorf("seat %d life %d → %d, want %d", i, lifeBefore[i], p.Life, want)
		}
	}
}

// The other direction: a printed creature that an effect turned into
// a noncreature before it died was not a creature when it died, though
// the graveyard card reads as one.
func TestZulaportIgnoresACreatureThatStoppedBeingOne(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Zulaport Cutthroat", "Creature — Human Rogue Ally", zulaportCutthroatOracle, false)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		if !g.RegisterScopedEffectForEffect(uuid.Nil, g.PinnedObjectsLocked(bear),
			[]game.Mod{game.RemoveTypesMod("Creature"), game.AddTypesMod("Artifact")},
			g.UntilEndOfTurnDuration(), "test — the bear becomes a noncreature artifact") {
			t.Fatal("setup: the effect registered nothing")
		}
		g.RecomputeLayersIfStaleLocked()
	})
	lifeBefore := lifeSnapshot(g)

	b18Kill(t, g, bear)
	if c, _ := g.LookupCardForEffect(bear); !c.IsCreature() {
		t.Fatal("setup: the graveyard card should read as the creature it prints — the point of the test")
	}
	for i, p := range g.Seats {
		if p.Life != lifeBefore[i] {
			t.Errorf("seat %d life %d → %d; a noncreature dying drains nobody", i, lifeBefore[i], p.Life)
		}
	}
}

// Midnight Reaper — "whenever a nontoken creature you control dies" —
// through the same read, for a crewed Vehicle sacrificed outside
// combat.
func TestMidnightReaperDrawsForASacrificedCrewedVehicle(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	pushCatalogPermanent(g, me.ID, "Midnight Reaper", "Creature — Zombie Knight", midnightReaperOracle, false)
	caravan := pushVehicleForTest(g, me.ID, "Cultivator's Caravan", cultivatorsCaravanOrcl, 5, 5)
	crewForTest(t, g, me.ID, caravan, pushCrewerForTest(g, me.ID, "Crewer", 3))
	handBefore, lifeBefore := me.Hand.Size(), me.Life

	g.WithWriteLock(func() {
		if err := g.SacrificePermanentForEffect(caravan); err != nil {
			t.Fatalf("SacrificePermanentForEffect: %v", err)
		}
	})
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != handBefore+1 || me.Life != lifeBefore-1 {
		t.Errorf("hand %d → %d, life %d → %d; want one card drawn and 1 life lost",
			handBefore, me.Hand.Size(), lifeBefore, me.Life)
	}
}

// Mahadi, Emporium Master counts "each creature that died this turn"
// off the per-turn tally, which reads the CR 603.10 snapshot as each
// death happens — the same types the event carries: a
// crewed Vehicle and an animated land that died are two creatures, an
// uncrewed Vehicle none.
func TestMahadiCountsACrewedVehicleAndAnAnimatedLand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b11Push(g, me.ID, "Mahadi, Emporium Master", "Legendary Creature — Devil", b11MahadiOracle, 3, 3)
	crewed := pushVehicleForTest(g, me.ID, "Cultivator's Caravan", cultivatorsCaravanOrcl, 5, 5)
	crewForTest(t, g, me.ID, crewed, pushCrewerForTest(g, me.ID, "Crewer", 3))
	idle := pushVehicleForTest(g, me.ID, "Cultivator's Caravan", cultivatorsCaravanOrcl, 5, 5)
	land := b12Permanent(g, me.ID, "Test Manland", "Land")
	animateLandForTest(t, g, land)

	b18Kill(t, g, crewed)
	b18Kill(t, g, idle)
	b18Kill(t, g, land)

	advanceToEndStepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	if n := countBattlefieldNamed(g, me.ID, "Treasure"); n != 2 {
		t.Errorf("a crewed Vehicle and an animated land died: %d Treasures, want 2", n)
	}
}

// Cruel Celebrant — "whenever this or another creature or planeswalker
// you control dies" — counts a crewed Vehicle through leftAsType.
func TestCruelCelebrantDrainsForACrewedVehicle(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	pushCatalogPermanent(g, me.ID, "Cruel Celebrant", "Creature — Vampire", b08CruelCelebrantOracle, false)
	caravan := pushVehicleForTest(g, me.ID, "Cultivator's Caravan", cultivatorsCaravanOrcl, 5, 5)
	crewForTest(t, g, me.ID, caravan, pushCrewerForTest(g, me.ID, "Crewer", 3))
	lifeBefore := me.Life

	b18Kill(t, g, caravan)
	if me.Life != lifeBefore+1 {
		t.Errorf("life %d → %d; a crewed Vehicle dying is a creature you controlled dying", lifeBefore, me.Life)
	}
}

// lastLTBOf is the newest EventLTB for cardID.
func lastLTBOf(t *testing.T, g *game.Game, cardID uuid.UUID) game.Event {
	t.Helper()
	for i := len(g.Events) - 1; i >= 0; i-- {
		if ev := g.Events[i]; ev.Kind == game.EventLTB && ev.CardID == cardID {
			return ev
		}
	}
	t.Fatalf("no EventLTB for %s", cardID)
	return game.Event{}
}

// Every dies / leaves-the-battlefield condition that tests a card
// type reads the last-known one (#1675). Three permanents whose types on
// the battlefield differ from the card left behind:
//
//   - a printed creature an effect made a noncreature artifact land,
//     which dies — the graveyard card is a creature and nothing else;
//   - a crewed Vehicle, which is bounced — the card in hand is an
//     artifact and nothing else;
//   - an animated land, which dies — the graveyard card is a land and
//     nothing else.
//
// Each condition answers for what the permanent was.
func TestLeavesTheBattlefieldConditionsReadLastKnownTypes(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	source := &game.Card{InstanceID: uuid.New(), Controller: me.ID, Owner: me.ID}

	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		if !g.RegisterScopedEffectForEffect(uuid.Nil, g.PinnedObjectsLocked(bear),
			[]game.Mod{game.RemoveTypesMod("Creature"), game.AddTypesMod("Artifact", "Land")},
			g.UntilEndOfTurnDuration(), "test — the bear becomes a noncreature artifact land") {
			t.Fatal("setup: the effect registered nothing")
		}
		g.RecomputeLayersIfStaleLocked()
	})
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(bear) })
	died := lastLTBOf(t, g, bear)

	caravan := pushVehicleForTest(g, me.ID, "Cultivator's Caravan", cultivatorsCaravanOrcl, 5, 5)
	crewForTest(t, g, me.ID, caravan, pushCrewerForTest(g, me.ID, "Crewer", 3))
	g.WithWriteLock(func() { _ = g.BounceToHandForEffect(caravan) })
	bounced := lastLTBOf(t, g, caravan)

	manland := b12Permanent(g, me.ID, "Test Manland", "Land")
	animateLandForTest(t, g, manland)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(manland) })
	landDied := lastLTBOf(t, g, manland)

	var none game.Characteristic
	_, diedAsCreature := diedCreature(died, g)
	b21Card, b21 := b21ArtifactOrCreatureYouControlDied(died, source, g)
	b21LandCard, b21Land := b21ArtifactOrCreatureYouControlDied(landDied, source, g)
	for _, c := range []struct {
		name      string
		got, want bool
	}{
		{"diedCreature (a noncreature died)", diedAsCreature, false},
		{"b21ArtifactOrCreatureYouControlDied", b21 && b21Card.InstanceID == bear, true},
		{"b23ArtifactPutIntoGraveyardFromBattlefield", b23ArtifactPutIntoGraveyardFromBattlefield(died, g), true},
		{"b10LandYouControlDied", b10LandYouControlDied(died, source, g), true},
		{"scrapTrawlerArtifactHitTheYard", scrapTrawlerArtifactHitTheYard(died, source, none, g), true},
		{"anotherArtifactOrCreaturePutIntoGraveyardFromBattlefield", anotherArtifactOrCreaturePutIntoGraveyardFromBattlefield(died, source, g), true},
		{"b21ArtifactOrCreatureYouControlDied (animated land died)", b21Land && b21LandCard.InstanceID == manland, true},
		{"anotherArtifactOrCreaturePutIntoGraveyardFromBattlefield (animated land died)", anotherArtifactOrCreaturePutIntoGraveyardFromBattlefield(landDied, source, g), true},
		{"creatureYouControlLeftWithoutDying (crewed Vehicle bounced)", creatureYouControlLeftWithoutDying(bounced, source, g), true},
		{"aCreatureYouControlLeft (crewed Vehicle bounced)", aCreatureYouControlLeft(bounced, source, none, g), true},
		{"aCreatureYouControlLeft (noncreature died)", aCreatureYouControlLeft(died, source, none, g), false},
	} {
		if c.got != c.want {
			t.Errorf("%s = %v, want %v", c.name, c.got, c.want)
		}
	}
}

// The Ozolith — "whenever a creature you control leaves the
// battlefield, if it had counters on it" — collects from a crewed
// Vehicle, which was a creature as it left.
func TestOzolithCollectsFromACrewedVehicle(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	oz := seedPermanentWithOracle(g, me.ID, "The Ozolith", "Legendary Artifact", ozolithOracle)
	caravan := pushVehicleForTest(g, me.ID, "Cultivator's Caravan", cultivatorsCaravanOrcl, 5, 5)
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == caravan {
			g.Battlefield.Cards[i].Counters = map[string]int{"+1/+1": 2}
		}
	}
	crewForTest(t, g, me.ID, caravan, pushCrewerForTest(g, me.ID, "Crewer", 3))

	b18Kill(t, g, caravan)
	if got := allCountersOn(t, g, oz); got["+1/+1"] != 2 {
		t.Errorf("The Ozolith's counters = %v, want +1/+1:2 from the crewed Vehicle", got)
	}
}
