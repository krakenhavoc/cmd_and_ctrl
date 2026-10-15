package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// top1000_sweep_d_test.go — the fourth slice of the missing top-1,000
// cards (#3654): Blind Obedience, Ramunap Excavator, Burgeoning,
// Archaeomancer's Map, Dragon Tempest, Blasphemous Edict, Second
// Harvest, Ozolith the Shattered Spire and Moraug, Fury of Akoum.

const (
	sweepDBlindObedienceOracle = "5d998c09-7d89-4265-ada4-6d80cbf56dae"
	sweepDRamunapOracle        = "4f819ba4-52ef-4fdd-8e4c-5ae3b2f44db5"
	sweepDBurgeoningOracle     = "88cf4c1a-5863-45fc-b76d-7ad33b175b2b"
	sweepDMapOracle            = "ef833546-f755-4ac5-867a-3926984e68a0"
	sweepDDragonTempestOracle  = "b9bacb46-fd1d-459b-81c2-d7d08fe73848"
	sweepDEdictOracle          = "6cf50e21-0f60-4e9a-b910-ebe1bad2a29e"
	sweepDSecondHarvestOracle  = "61a6341e-0d36-44b6-8026-9c87c8ef61cb"
	sweepDOzolithOracle        = "41fed659-237c-4d8e-ad31-d17fa0d3f764"
	sweepDMoraugOracle         = "6288c0cb-cf32-4ae9-abf0-104ec71654fa"
)

func TestSweepDBlindObedienceTapsOpposingArtifactsAndCreaturesOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Blind Obedience", "Enchantment", sweepDBlindObedienceOracle, false)

	g.WithWriteLock(func() {
		_ = g.CreateTokenForEffect(opp.ID, TokenCard("2/2 green Bear"), 1)
		_ = g.CreateTokenForEffect(opp.ID, TreasureToken(), 1)
		_ = g.CreateTokenForEffect(me.ID, TokenCard("2/2 green Bear"), 1)
	})
	tapped := func(owner uuid.UUID, treasure bool) (n, tappedN int) {
		for _, c := range g.Battlefield.Cards {
			if c.Controller != owner || (c.Name == "Blind Obedience") {
				continue
			}
			if c.HasSubtype("Treasure") != treasure {
				continue
			}
			n++
			if c.Tapped {
				tappedN++
			}
		}
		return
	}
	if n, tp := tapped(opp.ID, false); n != 1 || tp != 1 {
		t.Errorf("opposing creature: %d seen, %d tapped, want 1 and 1", n, tp)
	}
	if n, tp := tapped(opp.ID, true); n != 1 || tp != 1 {
		t.Errorf("opposing artifact: %d seen, %d tapped, want 1 and 1", n, tp)
	}
	if n, tp := tapped(me.ID, false); n != 1 || tp != 0 {
		t.Errorf("own creature: %d seen, %d tapped, want 1 and 0", n, tp)
	}
}

func TestSweepDBlindObedienceExtortsOnCast(t *testing.T) {
	spec, ok := Lookup(sweepDBlindObedienceOracle)
	if !ok {
		t.Fatal("Blind Obedience is not registered")
	}
	found := false
	for _, tr := range spec.Triggered {
		if tr.Keyword == KeywordExtort {
			found = true
		}
	}
	if !found {
		t.Error("Blind Obedience has no extort trigger")
	}
}

func TestSweepDRamunapExcavatorGrantsALandGraveyardPermission(t *testing.T) {
	spec, ok := Lookup(sweepDRamunapOracle)
	if !ok {
		t.Fatal("Ramunap Excavator is not registered")
	}
	if len(spec.CastPermissions) != 1 || spec.CastPermissions[0].Zone != game.ZoneGraveyard || !spec.CastPermissions[0].Filter.LandsOnly {
		t.Errorf("permissions %+v, want one lands-only graveyard permission", spec.CastPermissions)
	}
}

func TestSweepDBurgeoningPutsALandWhenAnOpponentPlaysOne(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Burgeoning", "Enchantment", sweepDBurgeoningOracle, false)
	land := handCardForTest(me, "Forest", "Basic Land — Forest", "")

	// My own land play does not trigger it.
	playLandFromHand(t, g, "Island", "")
	passPriorityAroundTable(t, g)
	if chooseCardsChoiceFor(g, me.ID) != nil {
		t.Fatal("Burgeoning triggered on its controller's own land play")
	}

	b13PlayAs(t, g, 1, "Mountain", "Basic Land — Mountain", "")
	passPriorityAroundTable(t, g)
	if chooseCardsChoiceFor(g, me.ID) == nil {
		t.Fatal("an opponent's land play did not offer the land from hand")
	}
	answerChooseCards(t, g, me.ID, land)
	if !onBattlefield(g, land) {
		t.Error("the chosen land did not enter the battlefield")
	}
}

func TestSweepDArchaeomancersMapOnlyTriggersWhenTheOpponentIsAhead(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Archaeomancer's Map", "Artifact", sweepDMapOracle, false)
	land := handCardForTest(me, "Forest", "Basic Land — Forest", "")

	// Level at one land each: not more than me, so no trigger.
	pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Plains", TypeLine: "Basic Land — Plains", Owner: me.ID, Controller: me.ID})
	b13PlayAs(t, g, 1, "Mountain", "Basic Land — Mountain", "")
	passPriorityAroundTable(t, g)
	if chooseCardsChoiceFor(g, me.ID) != nil {
		t.Fatal("triggered with the opponent's land count only equal to mine")
	}

	// Now they are ahead (two to one): it offers the land.
	pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Swamp", TypeLine: "Basic Land — Swamp", Owner: opp.ID, Controller: opp.ID})
	g.WithWriteLock(func() {
		_ = g.ReturnFromGraveyardUnderControlForEffect(b17GraveyardCard(opp, "Mountain", "Basic Land — Mountain", ""), game.ZoneBattlefield, opp.ID)
	})
	passPriorityAroundTable(t, g)
	if chooseCardsChoiceFor(g, me.ID) == nil {
		t.Fatal("an opponent who controls more lands did not trigger the Map")
	}
	answerChooseCards(t, g, me.ID, land)
	if !onBattlefield(g, land) {
		t.Error("the chosen land did not enter the battlefield")
	}
}

func TestSweepDArchaeomancersMapFetchesTwoPlains(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	p1 := pushLibraryCardForTest(me, game.Card{Name: "Plains", TypeLine: "Basic Land — Plains"})
	p2 := pushLibraryCardForTest(me, game.Card{Name: "Plains", TypeLine: "Basic Land — Plains"})
	p3 := pushLibraryCardForTest(me, game.Card{Name: "Plains", TypeLine: "Basic Land — Plains"})
	isl := pushLibraryCardForTest(me, game.Card{Name: "Island", TypeLine: "Basic Land — Island"})

	castCatalogSpell(t, g, "Archaeomancer's Map", "Artifact", sweepDMapOracle, nil)
	passPriorityAroundTable(t, g)
	takeSearchPrompts(t, g)
	passPriorityAroundTable(t, g)

	got := 0
	for _, id := range []uuid.UUID{p1, p2, p3} {
		if inHandOf(me, id) {
			got++
		}
	}
	if got != 2 {
		t.Errorf("%d Plains in hand, want 2", got)
	}
	if inHandOf(me, isl) {
		t.Error("fetched a non-Plains card")
	}
}

func TestSweepDDragonTempestHastesFlyersAndShootsForDragons(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	// A Dragon already in play counts toward X (placed first so its
	// own entry does not trigger the Tempest).
	pushCatalogPermanent(g, me.ID, "Old Dragon", "Creature — Dragon", "", false)
	pushCatalogPermanent(g, me.ID, "Dragon Tempest", "Enchantment", sweepDDragonTempestOracle, false)
	oppBefore := opp.Life

	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, TokenCard("4/4 red Dragon with flying"), 1) })
	// A flying Dragon trips both triggers: order them, then aim the shot.
	sweepDSettle(t, g, me.ID, opp.ID)

	if opp.Life != oppBefore-2 {
		t.Errorf("two Dragons should deal 2: %d → %d", oppBefore, opp.Life)
	}
	hasted := false
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.HasSubtype("Dragon") && c.IsToken() {
			v := s58p6View(t, g, c.InstanceID)
			hasted = game.HasKeyword(&v, "haste")
		}
	}
	if !hasted {
		t.Error("the flying Dragon did not gain haste")
	}

	// A non-Dragon flyer gets haste but shoots nothing.
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, TokenCard("1/1 white Bird with flying"), 1) })
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil {
		t.Error("a Bird triggered the damage ability")
	}
	bird := false
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.HasSubtype("Bird") {
			v := s58p6View(t, g, c.InstanceID)
			bird = game.HasKeyword(&v, "haste")
		}
	}
	if !bird {
		t.Error("the Bird did not gain haste")
	}
}

// sweepDSettle plays out a pair of simultaneous triggers: it aims any
// target prompt at `at`, answers the CR 603.3b ordering prompt, and
// passes priority until the stack is empty.
func sweepDSettle(t *testing.T, g *game.Game, who, at uuid.UUID) {
	t.Helper()
	for i := 0; i < 16; i++ {
		if latestPickTarget(g, who) != nil {
			pickPlayer(t, g, who, at)
			continue
		}
		var order *game.PendingChoice
		for _, c := range g.PendingChoices {
			if c != nil && c.Kind == game.PendingChoiceTriggerOrder && c.Chooser == who {
				order = c
			}
		}
		if order != nil {
			if err := g.ResolveTriggerOrder(order.ID, who, append([]uuid.UUID(nil), order.TriggerOrderIDs...)); err != nil {
				t.Fatalf("ResolveTriggerOrder: %v", err)
			}
			continue
		}
		if stackFullyEmpty(g) {
			return
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
}

func TestSweepDBlasphemousEdictChargesBOnlyWithThirteenCreatures(t *testing.T) {
	g := newCatalogGame(t)
	spec, ok := Lookup(sweepDEdictOracle)
	if !ok || len(spec.AlternativeCosts) != 1 {
		t.Fatal("Blasphemous Edict has no alternative cost")
	}
	cond := spec.AlternativeCosts[0].Condition
	me := g.Seats[0]
	if cond(g, me.ID) {
		t.Error("offered with no creatures")
	}
	for i := 0; i < 12; i++ {
		b12Creature(g, g.Seats[i%2].ID, "Bear", "Creature — Bear", 1, 1)
	}
	if cond(g, me.ID) {
		t.Error("offered with twelve creatures")
	}
	b12Creature(g, me.ID, "Bear", "Creature — Bear", 1, 1)
	if !cond(g, me.ID) {
		t.Error("not offered with thirteen creatures across both boards")
	}
}

func TestSweepDBlasphemousEdictMakesEveryPlayerSacrifice(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	var mine, theirs []uuid.UUID
	for i := 0; i < 14; i++ {
		mine = append(mine, b12Creature(g, me.ID, "Bear", "Creature — Bear", 1, 1))
	}
	for i := 0; i < 3; i++ {
		theirs = append(theirs, b12Creature(g, opp.ID, "Cub", "Creature — Bear", 1, 1))
	}
	castCatalogSpell(t, g, "Blasphemous Edict", "Sorcery", sweepDEdictOracle, nil)
	passPriorityAroundTable(t, g)
	// Each prompt takes one creature; keep answering with the first
	// of the chooser's creatures still in play until none is open.
	first := func(ids []uuid.UUID) uuid.UUID {
		for _, id := range ids {
			if onBattlefield(g, id) {
				return id
			}
		}
		return uuid.Nil
	}
	for i := 0; i < 40; i++ {
		if id := first(mine); id != uuid.Nil {
			answerSacrifice(t, g, me.ID, id)
		}
		if id := first(theirs); id != uuid.Nil {
			answerSacrifice(t, g, opp.ID, id)
		}
	}
	left := 0
	for _, id := range mine {
		if onBattlefield(g, id) {
			left++
		}
	}
	if left != 1 {
		t.Errorf("the caster has %d creatures left of 14, want 1", left)
	}
	for _, id := range theirs {
		if onBattlefield(g, id) {
			t.Error("an opponent with fewer than thirteen creatures kept one")
		}
	}
}

func TestSweepDSecondHarvestCopiesEachTokenYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	g.WithWriteLock(func() {
		_ = g.CreateTokenForEffect(me.ID, TokenCard("2/2 green Bear"), 2)
		_ = g.CreateTokenForEffect(me.ID, TreasureToken(), 1)
		_ = g.CreateTokenForEffect(opp.ID, TokenCard("2/2 green Bear"), 1)
	})
	b12Creature(g, me.ID, "Real Bear", "Creature — Bear", 2, 2)

	castCatalogSpell(t, g, "Second Harvest", "Instant", sweepDSecondHarvestOracle, nil)
	passPriorityAroundTable(t, g)

	count := func(owner uuid.UUID, name string) int {
		n := 0
		for _, c := range g.Battlefield.Cards {
			if c.Controller == owner && c.Name == name {
				n++
			}
		}
		return n
	}
	if got := count(me.ID, "Bear"); got != 4 {
		t.Errorf("Bear tokens: %d, want 4 (2 doubled)", got)
	}
	if got := countTreasures(g, me.ID); got != 2 {
		t.Errorf("Treasures: %d, want 2", got)
	}
	if got := count(me.ID, "Real Bear"); got != 1 {
		t.Errorf("a nontoken creature was copied: %d", got)
	}
	if got := count(opp.ID, "Bear"); got != 1 {
		t.Errorf("the opponent's token was copied: %d", got)
	}
}

func TestSweepDOzolithAddsACounterToYourArtifactsAndCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Ozolith, the Shattered Spire", "Legendary Artifact", sweepDOzolithOracle, false)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	rock := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Rock", TypeLine: "Artifact", Owner: me.ID, Controller: me.ID})
	enemy := b12Creature(g, opp.ID, "Enemy", "Creature — Bear", 2, 2)
	land := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})

	for _, id := range []uuid.UUID{bear, rock, enemy, land} {
		if err := g.AddCounter(id, "+1/+1", 1); err != nil {
			t.Fatalf("AddCounter: %v", err)
		}
	}
	for id, want := range map[uuid.UUID]int{bear: 2, rock: 2, enemy: 1, land: 1} {
		if got := allCountersOn(t, g, id)["+1/+1"]; got != want {
			t.Errorf("counters on %s: %d, want %d", id, got, want)
		}
	}
}

func TestSweepDMoraugPumpsByTimesAttacked(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Moraug, Fury of Akoum", "Legendary Creature — Minotaur Warrior", sweepDMoraugOracle, false)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)

	c := s58p6View(t, g, bear)
	if c.CurrentPower() != 2 {
		t.Errorf("before attacking power %d, want 2", c.CurrentPower())
	}
	advanceToDeclareAttackersOf(t, g, 0)
	if _, err := g.DeclareAttackers([]game.AttackDeclaration{{Attacker: bear, Target: g.Seats[1].ID}}); err != nil {
		t.Fatalf("DeclareAttackers: %v", err)
	}
	c = s58p6View(t, g, bear)
	if c.CurrentPower() != 3 {
		t.Errorf("after attacking once power %d, want 3", c.CurrentPower())
	}
}
