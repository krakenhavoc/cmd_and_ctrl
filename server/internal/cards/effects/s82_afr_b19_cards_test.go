package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_afr_b19_cards_test.go — AFR batch 19 (S82 #5499): Long Rest, Hunter's
// Mark, Split the Party, Sphere of Annihilation, Instrument of the Bards,
// Fey Steed, Ride the Avalanche and Hellish Rebuke. Meteor Swarm's tests
// are in meteor_swarm_test.go.

const (
	longRestOracle         = "fc8df5c8-fda6-454c-a1ac-21c6e3d75b9c"
	huntersMarkOracle      = "7def9bfe-6a59-4300-8134-fce7508dca38"
	splitThePartyOracle    = "9ad61841-9894-41da-bdd8-248c2ce21b00"
	sphereOfAnnihilOracle  = "b96399f9-c0ba-4507-be13-65cf716b7757"
	instrumentBardsOracle  = "d79755c4-3c61-4fc8-8eb7-fbdc23951748"
	feySteedOracle         = "18cced8b-9836-4e11-bb4e-69a971a33a50"
	rideTheAvalancheOracle = "d49c94fa-712a-47c0-b571-4db7e064a590"
	hellishRebukeOracle    = "5d307515-6ea1-4b1f-8e4b-49bfebfbdfcc"
)

// b19InHand puts a catalog card with a cost in the seat's hand.
func b19InHand(p *game.Player, name, typeLine, oracle, cost string, colors ...string) uuid.UUID {
	c := game.NewCard(name, p.ID)
	c.TypeLine, c.OracleID, c.ManaCost, c.Colors = typeLine, oracle, cost, colors
	p.Hand.PushTop(c)
	return c.InstanceID
}

// b19Graveyard puts a card with a mana cost into the seat's graveyard.
func b19Graveyard(p *game.Player, name, typeLine, cost string) uuid.UUID {
	c := game.NewCard(name, p.ID)
	c.TypeLine, c.ManaCost = typeLine, cost
	p.Graveyard.PushTop(c)
	return c.InstanceID
}

func b19Zone(g *game.Game, id uuid.UUID) game.ZoneKind {
	var kind game.ZoneKind
	g.WithWriteLock(func() {
		if z := g.FindCardZoneForEffect(id); z != nil {
			kind = z.Kind
		}
	})
	return kind
}

// --- Long Rest ---------------------------------------------------------------

func TestLongRestReturnsXCardsWithDifferentManaValuesAndExilesItself(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	a := b19Graveyard(me, "One", "Creature — Bear", "{G}")
	b := b19Graveyard(me, "Two", "Sorcery", "{1}{G}")
	same := b19Graveyard(me, "Also Two", "Instant", "{1}{R}")
	id := b19InHand(me, "Long Rest", "Sorcery", longRestOracle, "")
	advanceToMain(t, g)
	life := me.Life

	for name, picks := range map[string][]uuid.UUID{
		"same mana value": {b, same},
		"too few":         {a},
		"too many":        {a, b, same},
	} {
		err := g.CastSpell(me.ID, id, game.CastSpellParams{XValue: 2, Strict: true, Targets: cardRefs(picks...)})
		if !errors.Is(err, game.ErrInvalidParam) {
			t.Errorf("%s: CastSpell = %v, want ErrInvalidParam", name, err)
		}
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{XValue: 2, Strict: true, Targets: cardRefs(a, b)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if b19Zone(g, a) != game.ZoneHand || b19Zone(g, b) != game.ZoneHand {
		t.Error("the two targets should be in hand")
	}
	if b19Zone(g, same) != game.ZoneGraveyard {
		t.Error("the untargeted card should stay in the graveyard")
	}
	if b19Zone(g, id) != game.ZoneExile {
		t.Errorf("Long Rest should exile itself, it is in zone %v", b19Zone(g, id))
	}
	if me.Life != life {
		t.Errorf("two cards returned must not touch life: %d, want %d", me.Life, life)
	}
}

func TestLongRestWithEightCardsSetsLifeToTheStartingTotal(t *testing.T) {
	for _, start := range []int{12, 80} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		var picks []uuid.UUID
		for i := 0; i < 8; i++ {
			cost := "{" + string(rune('0'+i)) + "}"
			picks = append(picks, b19Graveyard(me, "Card", "Sorcery", cost))
		}
		id := b19InHand(me, "Long Rest", "Sorcery", longRestOracle, "")
		advanceToMain(t, g)
		me.Life = start
		if err := g.CastSpell(me.ID, id, game.CastSpellParams{XValue: 8, Strict: true, Targets: cardRefs(picks...)}); err != nil {
			t.Fatalf("cast: %v", err)
		}
		passPriorityAroundTable(t, g)
		if me.Life != 40 {
			t.Errorf("life %d with eight returned: %d, want the starting life total 40", start, me.Life)
		}
	}
}

func TestLongRestWithSevenCardsLeavesLifeAlone(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	var picks []uuid.UUID
	for i := 0; i < 7; i++ {
		picks = append(picks, b19Graveyard(me, "Card", "Sorcery", "{"+string(rune('0'+i))+"}"))
	}
	id := b19InHand(me, "Long Rest", "Sorcery", longRestOracle, "")
	advanceToMain(t, g)
	me.Life = 11
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{XValue: 7, Strict: true, Targets: cardRefs(picks...)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Life != 11 {
		t.Errorf("seven cards: life %d, want 11", me.Life)
	}
}

// --- Hunter's Mark -----------------------------------------------------------

func TestHuntersMarkPumpsThenBites(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Wall", 0, 5)
	b12PlayFromHand(t, g, "Hunter's Mark", "Instant", huntersMarkOracle,
		game.CastSpellParams{Targets: cardRefs(mine, theirs), Strict: true})
	passPriorityAroundTable(t, g)
	if p, tough := mrd2Size(t, g, mine); p != 3 || tough != 3 {
		t.Errorf("pumped creature is %d/%d, want 3/3", p, tough)
	}
	if d := e2Card(t, g, theirs).DamageMarked; d != 3 {
		t.Errorf("the bite dealt %d, want the pumped power 3", d)
	}
}

func TestHuntersMarkCostsThreeLessOnlyAgainstABlueOpposingPermanent(t *testing.T) {
	cast := func(blue bool, mana int) error {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		mine := b12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
		theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Wall", 0, 5)
		g.WithWriteLock(func() {
			for i := range g.Battlefield.Cards {
				if g.Battlefield.Cards[i].InstanceID == theirs && blue {
					g.Battlefield.Cards[i].Colors = []string{"U"}
				}
			}
		})
		id := b19InHand(me, "Hunter's Mark", "Instant", huntersMarkOracle, "{3}{G}", "G")
		advanceToMain(t, g)
		fillPoolColored(me, "G", mana)
		return g.CastSpell(me.ID, id, game.CastSpellParams{Targets: cardRefs(mine, theirs), Strict: true})
	}
	if err := cast(true, 1); err != nil {
		t.Errorf("blue target, {G} in the pool: %v, want it castable", err)
	}
	if err := cast(false, 1); err == nil {
		t.Error("a nonblue target must cost the full {3}{G}")
	}
	if err := cast(false, 4); err != nil {
		t.Errorf("nonblue target with four mana: %v", err)
	}
}

func TestHuntersMarkRefusesATargetYouDoNotOpposeOrControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	other := b12Creature(g, me.ID, "Other", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Wall", 0, 5)
	id := b19InHand(me, "Hunter's Mark", "Instant", huntersMarkOracle, "")
	advanceToMain(t, g)
	for name, picks := range map[string][]uuid.UUID{
		"both mine":       {mine, other},
		"first is theirs": {theirs, mine},
	} {
		if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: cardRefs(picks...), Strict: true}); err == nil {
			t.Errorf("%s: the cast should be refused", name)
		}
	}
}

// --- Split the Party ---------------------------------------------------------

func TestSplitThePartyReturnsHalfTheirCreaturesRoundedUp(t *testing.T) {
	for _, tc := range []struct{ creatures, want int }{{1, 1}, {2, 1}, {3, 2}, {4, 2}, {5, 3}} {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		var ids []uuid.UUID
		for i := 0; i < tc.creatures; i++ {
			ids = append(ids, b12Creature(g, opp.ID, "C", "Creature — Wall", 0, 3))
		}
		mine := b12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
		b12PlayFromHand(t, g, "Split the Party", "Sorcery", splitThePartyOracle,
			game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}})
		passPriorityAroundTable(t, g)
		c := chooseCardsChoiceFor(g, me.ID)
		if c == nil {
			t.Fatalf("%d creatures: the caster was not asked which to return", tc.creatures)
		}
		if hasID(c.ChooseCards, mine) {
			t.Errorf("%d creatures: my own creature was offered", tc.creatures)
		}
		answerChooseCards(t, g, me.ID, ids[:tc.want]...)
		passPriorityAroundTable(t, g)
		back := 0
		for _, id := range ids {
			if b19Zone(g, id) == game.ZoneHand {
				back++
			}
		}
		if back != tc.want {
			t.Errorf("%d creatures: %d returned, want %d", tc.creatures, back, tc.want)
		}
		if !g.Battlefield.Contains(mine) {
			t.Errorf("%d creatures: the caster's creature left", tc.creatures)
		}
	}
}

func TestSplitThePartyWithNoCreaturesAsksNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12PlayFromHand(t, g, "Split the Party", "Sorcery", splitThePartyOracle,
		game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}})
	passPriorityAroundTable(t, g)
	if chooseCardsChoiceFor(g, me.ID) != nil {
		t.Error("no creatures: nothing should be asked")
	}
}

// --- Sphere of Annihilation --------------------------------------------------

func TestSphereOfAnnihilationEntersWithXVoidCounters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := b12PlayFromHand(t, g, "Sphere of Annihilation", "Artifact", sphereOfAnnihilOracle,
		game.CastSpellParams{XValue: 3})
	passPriorityAroundTable(t, g)
	if got := countersOn(g, id, "void"); got != 3 {
		t.Errorf("void counters = %d, want 3", got)
	}
	_ = me
}

func TestSphereOfAnnihilationExilesByManaValueAtYourUpkeep(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	sphere := mrd2Catalog(g, me.ID, "Sphere of Annihilation", "Artifact", sphereOfAnnihilOracle, "{X}{B}", 0, 0)
	addCounters(t, g, sphere, "void", 2)
	smallMine := mrd2Catalog(g, me.ID, "Small Mine", "Creature — Bear", "", "{1}{G}", 2, 2)
	bigMine := mrd2Catalog(g, me.ID, "Big Mine", "Creature — Bear", "", "{2}{G}", 3, 3)
	smallTheirs := mrd2Catalog(g, opp.ID, "Small Theirs", "Creature — Wall", "", "{1}", 0, 3)
	walker := mrd2Catalog(g, opp.ID, "Walker", "Planeswalker — Test", "", "{1}{U}", 0, 0)
	land := mrd2Catalog(g, opp.ID, "Island", "Basic Land — Island", "", "", 0, 0)
	artifact := mrd2Catalog(g, opp.ID, "Rock", "Artifact", "", "{1}", 0, 0)
	gyCreature := b19Graveyard(opp, "Dead Small", "Creature — Wall", "{2}")
	gyBig := b19Graveyard(me, "Dead Big", "Creature — Wall", "{4}")
	gySpell := b19Graveyard(me, "Dead Spell", "Sorcery", "{1}")
	gyWalker := b19Graveyard(opp, "Dead Walker", "Planeswalker — Test", "{2}")

	advanceToUpkeepOf(t, g, 0)
	passPriorityAroundTable(t, g)

	for name, id := range map[string]uuid.UUID{"the Sphere": sphere, "small mine": smallMine, "small theirs": smallTheirs,
		"walker": walker, "graveyard creature": gyCreature, "graveyard walker": gyWalker} {
		if b19Zone(g, id) != game.ZoneExile {
			t.Errorf("%s should be exiled, it is in zone %v", name, b19Zone(g, id))
		}
	}
	if !g.Battlefield.Contains(bigMine) || !g.Battlefield.Contains(land) || !g.Battlefield.Contains(artifact) {
		t.Error("mana value 3, a land and an artifact must stay")
	}
	if b19Zone(g, gyBig) != game.ZoneGraveyard || b19Zone(g, gySpell) != game.ZoneGraveyard {
		t.Error("a mana value 4 creature card and a noncreature card must stay in the graveyard")
	}
}

func TestSphereOfAnnihilationOnlyTriggersAtItsControllersUpkeep(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	sphere := mrd2Catalog(g, me.ID, "Sphere of Annihilation", "Artifact", sphereOfAnnihilOracle, "{X}{B}", 0, 0)
	addCounters(t, g, sphere, "void", 5)
	bear := mrd2Catalog(g, opp.ID, "Bear", "Creature — Bear", "", "{1}{G}", 2, 2)
	advanceToUpkeepOf(t, g, 1)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(sphere) || !g.Battlefield.Contains(bear) {
		t.Error("an opponent's upkeep must not set it off")
	}
}

// --- Instrument of the Bards -------------------------------------------------

func b19Library(p *game.Player, name, typeLine, cost string) uuid.UUID {
	c := game.NewCard(name, p.ID)
	c.TypeLine, c.ManaCost = typeLine, cost
	p.Library.PushTop(c)
	return c.InstanceID
}

func TestInstrumentOfTheBardsMayAddAHarmonyCounterAtUpkeep(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	inst := mrd2Catalog(g, me.ID, "Instrument of the Bards", "Legendary Artifact", instrumentBardsOracle, "{G}", 0, 0)
	advanceToUpkeepOf(t, g, 0)
	b06AnswerAllTriggerPrompts(t, g, me.ID, true)
	if got := countersOn(g, inst, "harmony"); got != 1 {
		t.Fatalf("accepted: %d harmony counters, want 1", got)
	}
	advanceToUpkeepOf(t, g, 1)
	advanceToUpkeepOf(t, g, 0)
	b06AnswerAllTriggerPrompts(t, g, me.ID, false)
	if got := countersOn(g, inst, "harmony"); got != 1 {
		t.Errorf("declined: %d harmony counters, want still 1", got)
	}
}

func TestInstrumentOfTheBardsFindsAnExactManaValueAndTreasuresForALegend(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine string
		wantTreasure   bool
	}{{"Plain Bear", "Creature — Bear", false}, {"Hero", "Legendary Creature — Human", true}} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		advanceToMain(t, g)
		inst := mrd2Catalog(g, me.ID, "Instrument of the Bards", "Legendary Artifact", instrumentBardsOracle, "{G}", 0, 0)
		addCounters(t, g, inst, "harmony", 2)
		want := b19Library(me, tc.name, tc.typeLine, "{1}{G}")
		wrongValue := b19Library(me, "Three Drop", "Creature — Ape", "{2}{G}")
		notCreature := b19Library(me, "Two Spell", "Sorcery", "{1}{G}")
		fillPoolColored(me, "G", 4)
		if err := g.ActivateCatalogAbility(me.ID, inst, 0, game.ActivateAbilityParams{}); err != nil {
			t.Fatalf("activate: %v", err)
		}
		passPriorityAroundTable(t, g)
		c := searchChoiceFor(g, me.ID)
		if c == nil {
			t.Fatalf("%s: no search prompt", tc.name)
		}
		if searchOptionNamed(g, c, tc.name) == uuid.Nil || searchOptionNamed(g, c, "Three Drop") != uuid.Nil || searchOptionNamed(g, c, "Two Spell") != uuid.Nil {
			t.Fatalf("%s: the search must offer only the mana value 2 creature card", tc.name)
		}
		answerSearchByID(t, g, me.ID, want)
		passPriorityAroundTable(t, g)
		if b19Zone(g, want) != game.ZoneHand {
			t.Errorf("%s: the found card should be in hand", tc.name)
		}
		treasures := 0
		for _, bc := range g.Battlefield.Cards {
			if bc.Controller == me.ID && bc.HasSubtype("Treasure") {
				treasures++
			}
		}
		if (treasures == 1) != tc.wantTreasure || treasures > 1 {
			t.Errorf("%s: %d Treasures, want legendary=%v", tc.name, treasures, tc.wantTreasure)
		}
		_, _ = wrongValue, notCreature
	}
}

func TestInstrumentOfTheBardsNeedsItsFullCost(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	inst := mrd2Catalog(g, me.ID, "Instrument of the Bards", "Legendary Artifact", instrumentBardsOracle, "{G}", 0, 0)
	fillPoolColored(me, "G", 3)
	if err := g.ActivateCatalogAbility(me.ID, inst, 0, game.ActivateAbilityParams{}); err == nil {
		t.Error("three mana must not pay {3}{G}")
	}
}

// --- Fey Steed ---------------------------------------------------------------

func TestFeySteedAttackGrantsAnotherAttackerIndestructible(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	steed := dskB1Creature(g, me.ID, "Fey Steed", "Creature — Elk", feySteedOracle, 3, 3)
	other := b12Creature(g, me.ID, "Other Attacker", "Creature — Bear", 2, 2)
	idle := b12Creature(g, me.ID, "Stay Home", "Creature — Bear", 2, 2)
	declareAttack(t, g, opp.ID, steed, other)
	pick := pickTargetPrompt(t, g)
	for _, bad := range []uuid.UUID{steed, idle} {
		if hasID(pick.PickTargetCards, bad) {
			t.Errorf("%s was offered: the clause is another attacking creature", bad)
		}
	}
	answerPickTarget(t, g, other)
	passPriorityAroundTable(t, g)
	c := layeredCard(t, g, other)
	if !game.HasKeyword(&c, "indestructible") {
		t.Error("the target should be indestructible")
	}
	s := layeredCard(t, g, steed)
	if game.HasKeyword(&s, "indestructible") {
		t.Error("Fey Steed itself must not gain it")
	}
}

func TestFeySteedDrawsWhenAnOpponentTargetsYourCreatureOrPlaneswalker(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine string
		mine           bool
		want           int
	}{
		{"my creature", "Creature — Bear", true, 1},
		{"my planeswalker", "Planeswalker — Test", true, 1},
		{"my land", "Basic Land — Forest", true, 0},
		{"their creature", "Creature — Bear", false, 0},
	} {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1] // seat 0 is the active seat and casts
		dskB1Creature(g, opp.ID, "Fey Steed", "Creature — Elk", feySteedOracle, 3, 3)
		owner := opp.ID
		if !tc.mine {
			owner = me.ID
		}
		victim := mrd2Catalog(g, owner, "Victim", tc.typeLine, "", "{2}", 2, 5)
		if tc.typeLine == "Planeswalker — Test" {
			addCounters(t, g, victim, game.CounterLoyalty, 4)
		}
		advanceToMain(t, g)
		hand := len(opp.Hand.Cards)
		bolt := pushCatalogHandCard(me, "Lightning Bolt", "Instant", lightningBoltOracle)
		if err := g.CastSpell(me.ID, bolt, game.CastSpellParams{Targets: cardRefs(victim)}); err != nil {
			if tc.want == 0 {
				continue // an illegal target (a land) cannot trigger anything
			}
			t.Fatalf("%s: cast: %v", tc.name, err)
		}
		b06AnswerAllTriggerPrompts(t, g, opp.ID, true)
		if got := len(opp.Hand.Cards) - hand; got != tc.want {
			t.Errorf("%s: drew %d, want %d", tc.name, got, tc.want)
		}
	}
}

func TestFeySteedDoesNotDrawForYourOwnSpell(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	dskB1Creature(g, me.ID, "Fey Steed", "Creature — Elk", feySteedOracle, 3, 3)
	bear := seedBear(g, me.ID)
	advanceToMain(t, g)
	hand := len(me.Hand.Cards)
	bolt := pushCatalogHandCard(me, "Lightning Bolt", "Instant", lightningBoltOracle)
	hand++
	if err := g.CastSpell(me.ID, bolt, game.CastSpellParams{Targets: cardRefs(bear)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	b06AnswerAllTriggerPrompts(t, g, me.ID, true)
	if got := len(me.Hand.Cards); got != hand-1 {
		t.Errorf("hand %d, want %d: your own spell must not trigger it", got, hand-1)
	}
}

// --- Ride the Avalanche ------------------------------------------------------

func TestRideTheAvalancheCountersFromTheNextSpellsManaValueOnce(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := seedBear(g, me.ID)
	b12PlayFromHand(t, g, "Ride the Avalanche", "Instant", rideTheAvalancheOracle, game.CastSpellParams{})
	passPriorityAroundTable(t, g)

	first := b19InHand(me, "Three Drop", "Sorcery", "", "{2}{G}", "G")
	fillPoolColored(me, "G", 3)
	if err := g.CastSpell(me.ID, first, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast the next spell: %v", err)
	}
	pick := passUntilPickTarget(t, g, me.ID)
	if pick == nil {
		t.Fatal("the delayed trigger asked for no target")
	}
	answerPickTarget(t, g, bear)
	passPriorityAroundTable(t, g)
	if got := countersOn(g, bear, game.CounterPlusOne); got != 3 {
		t.Fatalf("counters = %d, want 3 (the spell's mana value)", got)
	}

	// It fires once: a second spell this turn does nothing.
	second := b19InHand(me, "Two Drop", "Sorcery", "", "{1}{G}", "G")
	fillPoolColored(me, "G", 2)
	if err := g.CastSpell(me.ID, second, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast the second spell: %v", err)
	}
	if latestPickTarget(g, me.ID) != nil {
		t.Error("the delayed trigger fired a second time")
	}
	passPriorityAroundTable(t, g)
	if got := countersOn(g, bear, game.CounterPlusOne); got != 3 {
		t.Errorf("counters = %d after the second spell, want still 3", got)
	}
}

func TestRideTheAvalancheTargetIsOptional(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := seedBear(g, me.ID)
	b12PlayFromHand(t, g, "Ride the Avalanche", "Instant", rideTheAvalancheOracle, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	spell := b19InHand(me, "Two Drop", "Sorcery", "", "{1}{G}", "G")
	fillPoolColored(me, "G", 2)
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	if pick := passUntilPickTarget(t, g, me.ID); pick != nil {
		b26DeclinePickTarget(t, g, me.ID)
	}
	passPriorityAroundTable(t, g)
	if got := countersOn(g, bear, game.CounterPlusOne); got != 0 {
		t.Errorf("declined: %d counters, want 0", got)
	}
}

func TestRideTheAvalancheTheNextSpellMayBeCastAsThoughItHadFlash(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12PlayFromHand(t, g, "Ride the Avalanche", "Instant", rideTheAvalancheOracle, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	// Passing around the table left the main phase with an empty stack; a
	// sorcery cast at the end of combat is only legal under the promise.
	advanceTo(t, g, game.StepEndCombat)
	sorcery := b19InHand(me, "Sorcery", "Sorcery", "", "")
	if err := g.CastSpell(me.ID, sorcery, game.CastSpellParams{}); err != nil {
		t.Errorf("a sorcery outside a main phase after Ride the Avalanche: %v", err)
	}
}

// --- Hellish Rebuke ----------------------------------------------------------

func TestHellishRebukePunishesDamageToTheCaster(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	biter := b12Creature(g, opp.ID, "Biter", "Creature — Wolf", 2, 2)
	land := b12Permanent(g, opp.ID, "Island", "Basic Land — Island")
	b12PlayFromHand(t, g, "Hellish Rebuke", "Instant", hellishRebukeOracle, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	life, myLife := opp.Life, me.Life

	g.WithWriteLock(func() {
		if err := g.DealDamageToPlayerForEffect(biter, me.ID, 2); err != nil {
			t.Fatalf("damage: %v", err)
		}
	})
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(biter) {
		t.Error("the creature that damaged the caster should be sacrificed")
	}
	if opp.Life != life-2 {
		t.Errorf("its controller's life = %d, want %d (they lose 2)", opp.Life, life-2)
	}
	if me.Life != myLife-2 {
		t.Errorf("the caster's life = %d, want only the damage: %d", me.Life, myLife-2)
	}
	if !g.Battlefield.Contains(land) {
		t.Error("an undamaging permanent was sacrificed")
	}
}

func TestHellishRebukeIgnoresOtherPlayersAndLaterPermanents(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	biter := b12Creature(g, opp.ID, "Biter", "Creature — Wolf", 2, 2)
	b12PlayFromHand(t, g, "Hellish Rebuke", "Instant", hellishRebukeOracle, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	late := b12Creature(g, opp.ID, "Late", "Creature — Wolf", 2, 2)
	mine := b12Creature(g, me.ID, "Mine", "Creature — Wolf", 2, 2)
	life := opp.Life

	g.WithWriteLock(func() {
		if err := g.DealDamageToPlayerForEffect(biter, third.ID, 2); err != nil {
			t.Fatalf("damage to the third player: %v", err)
		}
		if err := g.DealDamageToPlayerForEffect(late, me.ID, 2); err != nil {
			t.Fatalf("damage from the late permanent: %v", err)
		}
		if err := g.DealDamageToPlayerForEffect(mine, opp.ID, 2); err != nil {
			t.Fatalf("damage from my own permanent: %v", err)
		}
	})
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(biter) || !g.Battlefield.Contains(late) || !g.Battlefield.Contains(mine) {
		t.Error("only a permanent that was an opponent's at resolution and damaged the caster is sacrificed")
	}
	if opp.Life != life-2 {
		t.Errorf("opponent life %d, want %d (only my permanent's damage)", opp.Life, life-2)
	}
}

func TestHellishRebukeEndsWithTheTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	biter := b12Creature(g, opp.ID, "Biter", "Creature — Wolf", 2, 2)
	b12PlayFromHand(t, g, "Hellish Rebuke", "Instant", hellishRebukeOracle, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	advanceToUpkeepOf(t, g, 2)
	g.WithWriteLock(func() {
		if err := g.DealDamageToPlayerForEffect(biter, me.ID, 1); err != nil {
			t.Fatalf("damage: %v", err)
		}
	})
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(biter) {
		t.Error("the granted ability should be gone after the turn ended")
	}
}
