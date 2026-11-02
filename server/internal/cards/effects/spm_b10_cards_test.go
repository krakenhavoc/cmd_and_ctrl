package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// spm_b10_cards_test.go — Marvel's Spider-Man, batch 10 (#4460). Each test
// plays the real catalog entry and asserts what its text says.

const (
	spmB10GwenOracle      = "2624b5c5-ba59-4e20-8e66-e463268366e4"
	spmB10SpotOracle      = "a5dfa7ad-9637-4da8-9ed2-0ac863617d9a"
	spmB10GoblinOracle    = "b5b43d01-fce6-4a00-9c19-7a7e2a09d833"
	spmB10CampusOracle    = "98e547de-b963-4ee4-9a08-67bae010734b"
	spmB10RetreatOracle   = "18290c4b-cff6-4ee5-a487-fe0d5706d3de"
	spmB10HungerOracle    = "8388faa0-f64e-4b01-836e-617f99b808a4"
	spmB10WrathOracle     = "1c7c1867-20cb-4035-bb2a-42760edf0a55"
	spmB10VultureOracle   = "0409765a-bbb4-465e-99b3-3687bffc3cf9"
	spmB10WallCrawlOracle = "4a09536e-dfc1-4b15-adc1-07ef866ebda0"
	spmB10WebOracle       = "33d556c5-1bb4-49bb-88c2-627311eb8c9b"

	spmB10SpotTypeLine   = "Legendary Creature — Human Scientist Villain"
	spmB10GoblinTypeLine = "Legendary Creature — Goblin Villain"
)

func TestSpmB10EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		spmB10GwenOracle:      "The Death of Gwen Stacy",
		spmB10SpotOracle:      "The Spot, Living Portal",
		spmB10GoblinOracle:    "Ultimate Green Goblin",
		spmB10CampusOracle:    "University Campus",
		spmB10RetreatOracle:   "Urban Retreat",
		spmB10HungerOracle:    "Venom's Hunger",
		spmB10WrathOracle:     "Villainous Wrath",
		spmB10VultureOracle:   "Vulture, Scheming Scavenger",
		spmB10WallCrawlOracle: "Wall Crawl",
		spmB10WebOracle:       "Web of Life and Destiny",
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

// spmB10SetHand replaces a player's hand with n plain cards, returning ids.
func spmB10SetHand(g *game.Game, p *game.Player, n int) []uuid.UUID {
	var ids []uuid.UUID
	g.WithWriteLock(func() {
		p.Hand.Cards = nil
		for i := 0; i < n; i++ {
			id := uuid.New()
			ids = append(ids, id)
			p.Hand.PushTop(game.Card{
				InstanceID: id, Name: "Filler", TypeLine: "Sorcery", Owner: p.ID, Controller: p.ID,
				KnownBy: map[uuid.UUID]bool{p.ID: true},
			})
		}
	})
	return ids
}

// spmB10GraveyardCard puts a card with the type line into p's graveyard.
func spmB10GraveyardCard(g *game.Game, p *game.Player, name, typeLine string) uuid.UUID {
	id := uuid.New()
	g.WithWriteLock(func() {
		p.Graveyard.PushTop(game.Card{
			InstanceID: id, Name: name, TypeLine: typeLine, Power: 2, Toughness: 2,
			Owner: p.ID, Controller: p.ID,
		})
	})
	return id
}

// --- The Death of Gwen Stacy ---

func TestSpmB10GwenStacyChapterOneDestroysATargetCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	victim := pushVanillaCreature(g, opp.ID, "Victim", 3, 3)
	mine := pushVanillaCreature(g, me.ID, "Mine", 1, 1)
	land := b12Permanent(g, opp.ID, "Forest", "Basic Land — Forest")

	saga := castCatalogSpell(t, g, "The Death of Gwen Stacy", "Enchantment — Saga", spmB10GwenOracle, nil)
	passPriorityAroundTable(t, g)
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no chapter I target prompt")
	}
	if !hasID(p.PickTargetCards, victim) || !hasID(p.PickTargetCards, mine) || hasID(p.PickTargetCards, land) {
		t.Fatalf("chapter I offered %v; want every creature and no land", p.PickTargetCards)
	}
	pickCard(t, g, me.ID, victim)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Error("chapter I did not destroy the creature")
	}
	if !g.Battlefield.Contains(mine) || !g.Battlefield.Contains(saga) {
		t.Error("only the target should have gone")
	}
}

func spmB10GwenAtChapter(t *testing.T, g *game.Game, me *game.Player, lore int) uuid.UUID {
	t.Helper()
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "The Death of Gwen Stacy", TypeLine: "Enchantment — Saga",
		OracleID: spmB10GwenOracle, Owner: me.ID, Controller: me.ID,
		Counters: map[string]int{game.CounterLore: lore},
	})
}

func TestSpmB10GwenStacyChapterTwoDiscardOrLoseThreeInTurnOrder(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	n := len(g.Seats)
	a, b, c, d := g.Seats[seat], g.Seats[(seat+1)%n], g.Seats[(seat+2)%n], g.Seats[(seat+3)%n]
	saga := spmB10GwenAtChapter(t, g, a, 1)

	advanceToPrecombatMainOf(t, g, seat)
	ha := spmB10SetHand(g, a, 2)
	spmB10SetHand(g, b, 2)
	spmB10SetHand(g, c, 0)
	hd := spmB10SetHand(g, d, 1)
	lives := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		lives[p.ID] = p.Life
	}
	passPriorityAroundTable(t, g)

	answerDiscard(t, g, a.ID, ha[0])
	// Nobody has lost anything while the table is still deciding.
	for _, p := range g.Seats {
		if p.Life != lives[p.ID] {
			t.Fatalf("seat %s lost life before every answer was in", p.ID)
		}
	}
	answerDiscard(t, g, b.ID) // declines
	answerDiscard(t, g, d.ID, hd[0])

	want := map[uuid.UUID]int{a.ID: 0, b.ID: -3, c.ID: -3, d.ID: 0}
	for _, p := range g.Seats {
		if got := p.Life - lives[p.ID]; got != want[p.ID] {
			t.Errorf("seat %s life change %d, want %d", p.ID, got, want[p.ID])
		}
	}
	if !a.Graveyard.Contains(ha[0]) || !d.Graveyard.Contains(hd[0]) {
		t.Error("the discarded cards are not in their owners' graveyards")
	}
	if b.Hand.Size() != 2 {
		t.Errorf("the declining player's hand changed: %d", b.Hand.Size())
	}
	if loreCountersOn(g, saga) != 2 {
		t.Errorf("lore = %d, want 2", loreCountersOn(g, saga))
	}
}

func TestSpmB10GwenStacyChapterThreeExilesTheChosenGraveyardsThenIsSacrificed(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	n := len(g.Seats)
	a, b, c, d := g.Seats[seat], g.Seats[(seat+1)%n], g.Seats[(seat+2)%n], g.Seats[(seat+3)%n]
	saga := spmB10GwenAtChapter(t, g, a, 2)
	ga := spmB10GraveyardCard(g, a, "A1", "Creature — Bear")
	gb := spmB10GraveyardCard(g, b, "B1", "Creature — Bear")
	gc := spmB10GraveyardCard(g, c, "C1", "Instant")
	gd := spmB10GraveyardCard(g, d, "D1", "Creature — Bear")

	advanceToPrecombatMainOf(t, g, seat)
	passPriorityAroundTable(t, g)
	p := latestPickTarget(g, a.ID)
	if p == nil {
		t.Fatal("no chapter III target prompt")
	}
	if err := g.ResolvePickTargets(p.ID, a.ID, []game.TargetRef{
		{Kind: game.TargetPlayer, ID: b.ID}, {Kind: game.TargetPlayer, ID: c.ID},
	}); err != nil {
		t.Fatalf("ResolvePickTargets: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(gb) || !g.Exile.Contains(gc) || b.Graveyard.Contains(gb) || c.Graveyard.Contains(gc) {
		t.Error("the chosen players' graveyards were not exiled")
	}
	if !a.Graveyard.Contains(ga) || !d.Graveyard.Contains(gd) {
		t.Error("an unchosen graveyard was exiled")
	}
	if loreCountersOn(g, saga) >= 0 {
		t.Error("the Saga was not sacrificed after chapter III")
	}
}

func TestSpmB10GwenStacyChapterThreeMayTargetNoOne(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	a, b := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
	saga := spmB10GwenAtChapter(t, g, a, 2)
	gb := spmB10GraveyardCard(g, b, "B1", "Creature — Bear")

	advanceToPrecombatMainOf(t, g, seat)
	passPriorityAroundTable(t, g)
	if p := latestPickTarget(g, a.ID); p != nil {
		if err := g.ResolvePickTargets(p.ID, a.ID, nil); err != nil {
			t.Fatalf("choosing no targets: %v", err)
		}
	}
	passPriorityAroundTable(t, g)
	if !b.Graveyard.Contains(gb) {
		t.Error("a graveyard was exiled with no target")
	}
	if loreCountersOn(g, saga) >= 0 {
		t.Error("the Saga was not sacrificed after chapter III")
	}
}

// --- The Spot, Living Portal ---

func spmB10CastSpot(t *testing.T, g *game.Game) uuid.UUID {
	t.Helper()
	return castHoldCreature(t, g, "The Spot, Living Portal", spmB10SpotTypeLine, spmB10SpotOracle, 4, 4)
}

func TestSpmB10SpotExilesAPermanentAndAGraveyardCardThenReturnsThemWhenItDies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	victim := pushVanillaCreature(g, opp.ID, "Victim", 3, 3)
	land := b12Permanent(g, opp.ID, "Forest", "Basic Land — Forest")
	gyCreature := spmB10GraveyardCard(g, opp, "Gy Bear", "Creature — Bear")
	gyInstant := spmB10GraveyardCard(g, opp, "Gy Bolt", "Instant")
	otherExiled := uuid.New()
	g.WithWriteLock(func() {
		g.Exile.PushTop(game.Card{InstanceID: otherExiled, Name: "Elsewhere", TypeLine: "Creature — Bear", Owner: opp.ID, Controller: opp.ID})
	})

	spot := spmB10CastSpot(t, g)
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no first target prompt")
	}
	if hasID(p.PickTargetCards, land) || !hasID(p.PickTargetCards, victim) {
		t.Fatalf("first clause offered %v; want nonland permanents only", p.PickTargetCards)
	}
	pickCard(t, g, me.ID, victim)
	p = latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no second target prompt")
	}
	if !hasID(p.PickTargetCards, gyCreature) || hasID(p.PickTargetCards, gyInstant) {
		t.Fatalf("second clause offered %v; want the permanent card and not the instant", p.PickTargetCards)
	}
	pickCard(t, g, me.ID, gyCreature)
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(victim) || !g.Exile.Contains(gyCreature) {
		t.Fatal("the chosen permanent and graveyard card were not both exiled")
	}

	g.WithWriteLock(func() { g.DestroyPermanentsForEffect([]uuid.UUID{spot}) })
	passPriorityAroundTable(t, g)
	if me.Library.Size() == 0 || me.Library.Cards[0].InstanceID != spot {
		t.Error("The Spot is not on the bottom of his owner's library")
	}
	if me.Graveyard.Contains(spot) {
		t.Error("The Spot stayed in the graveyard")
	}
	if !opp.Hand.Contains(victim) || !opp.Hand.Contains(gyCreature) {
		t.Error("the exiled cards did not return to their owners' hands")
	}
	if !g.Exile.Contains(otherExiled) || opp.Hand.Contains(otherExiled) {
		t.Error("a card exiled by something else came back")
	}
	if !opp.Graveyard.Contains(gyInstant) {
		t.Error("the unchosen graveyard card moved")
	}
}

func TestSpmB10SpotWithNothingExiledStillGoesToTheBottom(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	spot := spmB10CastSpot(t, g)
	// Both clauses are "up to one": choose neither.
	for p := latestPickTarget(g, me.ID); p != nil; p = latestPickTarget(g, me.ID) {
		if err := g.ResolvePickTargets(p.ID, me.ID, nil); err != nil {
			t.Fatalf("choosing no target: %v", err)
		}
	}
	passPriorityAroundTable(t, g)
	g.WithWriteLock(func() { g.DestroyPermanentsForEffect([]uuid.UUID{spot}) })
	passPriorityAroundTable(t, g)
	if me.Library.Size() == 0 || me.Library.Cards[0].InstanceID != spot {
		t.Error("The Spot is not on the bottom of his owner's library")
	}
}

func TestSpmB10SpotLeavingByBounceDoesNotReturnTheExiledCard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	victim := pushVanillaCreature(g, opp.ID, "Victim", 3, 3)
	spot := spmB10CastSpot(t, g)
	pickCard(t, g, me.ID, victim)
	if p := latestPickTarget(g, me.ID); p != nil {
		if err := g.ResolvePickTargets(p.ID, me.ID, nil); err != nil {
			t.Fatalf("second clause: %v", err)
		}
	}
	passPriorityAroundTable(t, g)
	// He leaves by bounce, not by dying: nothing returns the exiled card.
	g.WithWriteLock(func() { _ = g.BounceToHandForEffect(spot) })
	if !g.Exile.Contains(victim) {
		t.Fatal("the exiled card should stay exiled")
	}
}

// --- Ultimate Green Goblin ---

func TestSpmB10GoblinUpkeepDiscardsThenCreatesATreasure(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	b12Push(g, opp.ID, "Ultimate Green Goblin", spmB10GoblinTypeLine, spmB10GoblinOracle, 5, 4)
	hand := spmB10SetHand(g, opp, 3)
	for i := 0; i < 80 && !(g.Turn.Step == game.StepUpkeep && g.Turn.ActiveSeat == 1); i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	passPriorityAroundTable(t, g)
	if countBattlefieldByName(g, "Treasure") != 0 {
		t.Fatal("the Treasure came before the discard")
	}
	answerDiscard(t, g, opp.ID, hand[1])
	if !opp.Graveyard.Contains(hand[1]) {
		t.Error("the chosen card was not discarded")
	}
	if got := countBattlefieldByName(g, "Treasure"); got != 1 {
		t.Errorf("Treasures = %d, want 1", got)
	}
	if opp.Hand.Size() != 2 {
		t.Errorf("hand = %d, want 2 (one discarded)", opp.Hand.Size())
	}
}

func TestSpmB10GoblinWithAnEmptyHandStillMakesATreasure(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	b12Push(g, opp.ID, "Ultimate Green Goblin", spmB10GoblinTypeLine, spmB10GoblinOracle, 5, 4)
	for i := 0; i < 80 && !(g.Turn.Step == game.StepUpkeep && g.Turn.ActiveSeat == 1); i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	spmB10SetHand(g, opp, 0)
	passPriorityAroundTable(t, g)
	if got := countBattlefieldByName(g, "Treasure"); got != 1 {
		t.Errorf("Treasures = %d, want 1 with nothing to discard", got)
	}
}

func TestSpmB10GoblinMayhemCastsFromTheGraveyardForTheMayhemCost(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{
			InstanceID: id, Name: "Ultimate Green Goblin", TypeLine: spmB10GoblinTypeLine,
			ManaCost: "{1}{B/R}{B/R}", OracleID: spmB10GoblinOracle, Power: 5, Toughness: 4,
			Owner: me.ID, Controller: me.ID, KnownBy: map[uuid.UUID]bool{me.ID: true},
		})
	})
	advanceToMain(t, g)
	mayhemDiscard(t, g, me, id)

	floatForTest(g, me, "CB")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{FromZone: "graveyard", AlternativeCost: "mayhem", Strict: true}); err == nil {
		t.Fatal("mayhem {2}{B/R} paid with two mana")
	}
	if !me.Graveyard.Contains(id) {
		t.Fatal("a refused cast moved the card")
	}
	floatForTest(g, me, "C")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{FromZone: "graveyard", AlternativeCost: "mayhem", Strict: true}); err != nil {
		t.Fatalf("mayhem cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(id) {
		t.Error("the Goblin did not enter")
	}
}

// --- University Campus ---

func TestSpmB10CampusEntersTappedMakesWOrUAndSurveilsForFour(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedLibrary(me, "Top", "Next")
	id := playLandFromHand(t, g, "University Campus", spmB10CampusOracle)
	passPriorityAroundTable(t, g)
	card, ok := battlefieldCard(g, id)
	if !ok || !card.Tapped {
		t.Fatalf("on battlefield %v, tapped %v; want it to enter tapped", ok, card.Tapped)
	}
	untapForTest(g, id)
	if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{Colors: []string{"B"}}); err == nil {
		t.Fatal("the land made {B}")
	}
	if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{Colors: []string{"U"}}); err != nil {
		t.Fatalf("{T}: Add {U}: %v", err)
	}
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })

	untapForTest(g, id)
	floatForTest(g, me, "CCC")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("surveil activated with three mana")
	}
	floatForTest(g, me, "C")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("{4},{T}: surveil 1: %v", err)
	}
	passPriorityAroundTable(t, g)
	c := surveilChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no surveil prompt")
	}
	if len(c.ScryCards) != 1 {
		t.Fatalf("surveil looked at %d cards, want 1", len(c.ScryCards))
	}
	top := c.ScryCards[0]
	if err := g.ResolveSurveil(c.ID, me.ID, []uuid.UUID{top}, nil); err != nil {
		t.Fatalf("ResolveSurveil: %v", err)
	}
	if !me.Graveyard.Contains(top) {
		t.Error("the surveilled card did not go to the graveyard")
	}
	if card, _ := battlefieldCard(g, id); !card.Tapped {
		t.Error("the {T} cost did not tap the land")
	}
}

// --- Urban Retreat ---

func spmB10RetreatInHand(g *game.Game, me *game.Player) uuid.UUID {
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{
			InstanceID: id, Name: "Urban Retreat", TypeLine: "Land", OracleID: spmB10RetreatOracle,
			Owner: me.ID, Controller: me.ID, KnownBy: map[uuid.UUID]bool{me.ID: true},
		})
	})
	return id
}

func TestSpmB10RetreatPutsItselfOntoTheBattlefieldTappedByReturningATappedCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	land := spmB10RetreatInHand(g, me)
	tapped := pushVanillaCreature(g, me.ID, "Tapped Bear", 2, 2)
	untapped := pushVanillaCreature(g, me.ID, "Untapped Bear", 2, 2)
	tapForTest(t, g, tapped)

	floatForTest(g, me, "CC")
	if err := g.ActivateCatalogAbility(me.ID, land, 0, game.ActivateAbilityParams{Strict: true, ReturnIDs: []uuid.UUID{untapped}}); err == nil {
		t.Fatal("returned an untapped creature for the cost")
	}
	if g.Battlefield.Contains(land) || !g.Battlefield.Contains(untapped) {
		t.Fatal("a refused activation changed the board")
	}
	if err := g.ActivateCatalogAbility(me.ID, land, 0, game.ActivateAbilityParams{Strict: true, ReturnIDs: []uuid.UUID{tapped}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	card, ok := battlefieldCard(g, land)
	if !ok {
		t.Fatal("Urban Retreat did not enter")
	}
	if !card.Tapped {
		t.Error("Urban Retreat should enter tapped")
	}
	if !me.Hand.Contains(tapped) || g.Battlefield.Contains(tapped) {
		t.Error("the tapped creature was not returned to its owner's hand")
	}
}

func TestSpmB10RetreatRefusesWithoutTheManaOrOutsideSorcerySpeed(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	land := spmB10RetreatInHand(g, me)
	tapped := pushVanillaCreature(g, me.ID, "Tapped Bear", 2, 2)
	tapForTest(t, g, tapped)
	floatForTest(g, me, "C")
	if err := g.ActivateCatalogAbility(me.ID, land, 0, game.ActivateAbilityParams{Strict: true, ReturnIDs: []uuid.UUID{tapped}}); err == nil {
		t.Fatal("activated {2} with one mana")
	}
	floatForTest(g, me, "C")
	advanceTo(t, g, game.StepBeginCombat)
	if err := g.ActivateCatalogAbility(me.ID, land, 0, game.ActivateAbilityParams{Strict: true, ReturnIDs: []uuid.UUID{tapped}}); err == nil {
		t.Fatal("activated at a non-sorcery time")
	}
	if !me.Hand.Contains(land) {
		t.Error("the land left the hand")
	}
}

func TestSpmB10RetreatMakesGWOrU(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Urban Retreat", TypeLine: "Land", OracleID: spmB10RetreatOracle,
		Owner: me.ID, Controller: me.ID,
	})
	for _, c := range []string{"G", "W", "U"} {
		untapForTest(g, id)
		if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{Colors: []string{c}}); err != nil {
			t.Fatalf("{T}: Add {%s}: %v", c, err)
		}
		g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	}
	untapForTest(g, id)
	if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{Colors: []string{"R"}}); err == nil {
		t.Fatal("the land made {R}")
	}
}

// --- Venom's Hunger ---

func TestSpmB10HungerCostsTwoLessWithAVillain(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	if got := selfPricedMV(t, g, me, spmB10HungerOracle, "Sorcery", "{4}{B}"); got != 5 {
		t.Errorf("without a Villain: %d, want 5", got)
	}
	b12Creature(g, g.Seats[1].ID, "Their Villain", "Creature — Human Villain", 2, 2)
	if got := selfPricedMV(t, g, me, spmB10HungerOracle, "Sorcery", "{4}{B}"); got != 5 {
		t.Errorf("an opponent's Villain: %d, want 5", got)
	}
	b12Creature(g, me.ID, "My Villain", "Creature — Human Villain", 2, 2)
	if got := selfPricedMV(t, g, me, spmB10HungerOracle, "Sorcery", "{4}{B}"); got != 3 {
		t.Errorf("with a Villain: %d, want 3", got)
	}
}

func spmB10HungerHand(g *game.Game, me *game.Player) uuid.UUID {
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{
			InstanceID: id, Name: "Venom's Hunger", TypeLine: "Sorcery", ManaCost: "{4}{B}",
			OracleID: spmB10HungerOracle, Owner: me.ID, Controller: me.ID,
			KnownBy: map[uuid.UUID]bool{me.ID: true},
		})
	})
	return id
}

func TestSpmB10HungerDestroysTheCreatureAndGainsTwoAtTheDiscountedPrice(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	advanceToMain(t, g)
	victim := pushVanillaCreature(g, opp.ID, "Victim", 3, 3)
	spell := spmB10HungerHand(g, me)
	target := []game.TargetRef{{Kind: game.TargetCard, ID: victim}}

	floatForTest(g, me, "CCB")
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, Targets: target}); err == nil {
		t.Fatal("cast for {2}{B} without a Villain")
	}
	b12Creature(g, me.ID, "My Villain", "Creature — Human Villain", 2, 2)
	life := me.Life
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, Targets: target}); err != nil {
		t.Fatalf("cast for {2}{B} with a Villain: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Error("the creature survived")
	}
	if me.Life != life+2 {
		t.Errorf("life %d -> %d, want +2", life, me.Life)
	}
}

func TestSpmB10HungerCannotTargetANoncreature(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	land := b12Permanent(g, opp.ID, "Forest", "Basic Land — Forest")
	err := castCatalogSpellErr(t, g, "Venom's Hunger", "Sorcery", spmB10HungerOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: land}})
	if err == nil {
		t.Fatal("targeted a land")
	}
}

// --- Villainous Wrath ---

func TestSpmB10WrathDrainsForTheirCreaturesThenDestroysEveryCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, other := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)], g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)]
	theirs := []uuid.UUID{
		pushVanillaCreature(g, opp.ID, "T1", 1, 1), pushVanillaCreature(g, opp.ID, "T2", 1, 1), pushVanillaCreature(g, opp.ID, "T3", 1, 1),
	}
	mine := pushVanillaCreature(g, me.ID, "Mine", 2, 2)
	third := pushVanillaCreature(g, other.ID, "Third", 2, 2)
	land := b12Permanent(g, opp.ID, "Forest", "Basic Land — Forest")
	oppLife, otherLife, myLife := opp.Life, other.Life, me.Life

	castCatalogSpell(t, g, "Villainous Wrath", "Sorcery", spmB10WrathOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	passPriorityAroundTable(t, g)
	if opp.Life != oppLife-3 {
		t.Errorf("target opponent life %d -> %d, want -3", oppLife, opp.Life)
	}
	if other.Life != otherLife || me.Life != myLife {
		t.Error("only the target opponent loses life")
	}
	for _, id := range append(theirs, mine, third) {
		if g.Battlefield.Contains(id) {
			t.Errorf("creature %s survived", id)
		}
	}
	if !g.Battlefield.Contains(land) {
		t.Error("the land was destroyed")
	}
}

func TestSpmB10WrathRefusesAYouTarget(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	err := castCatalogSpellErr(t, g, "Villainous Wrath", "Sorcery", spmB10WrathOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}})
	if err == nil {
		t.Fatal("targeted yourself")
	}
}

// --- Vulture, Scheming Scavenger ---

func TestSpmB10VultureGivesOtherVillainsFlyingUntilEndOfTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	vulture := b12Push(g, me.ID, "Vulture, Scheming Scavenger", "Legendary Creature — Human Artificer Villain", spmB10VultureOracle, 4, 6)
	assertKeywords(t, g, vulture, "flying")
	villain := b12Creature(g, me.ID, "My Villain", "Creature — Human Villain", 2, 2)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Villain", "Creature — Human Villain", 2, 2)

	declareAttack(t, g, opp.ID, vulture)
	passPriorityAroundTable(t, g)
	if !hasEffectiveKeyword(t, g, villain, "flying") {
		t.Error("another Villain you control should gain flying")
	}
	if hasEffectiveKeyword(t, g, bear, "flying") || hasEffectiveKeyword(t, g, theirs, "flying") {
		t.Error("a non-Villain or an opponent's Villain gained flying")
	}
	advanceToPrecombatMainOf(t, g, 1)
	if hasEffectiveKeyword(t, g, villain, "flying") {
		t.Error("flying should end with the turn")
	}
}

func TestSpmB10VultureDoesNotTriggerWhenAnotherCreatureAttacks(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	b12Push(g, me.ID, "Vulture, Scheming Scavenger", "Legendary Creature — Human Artificer Villain", spmB10VultureOracle, 4, 6)
	villain := b12Creature(g, me.ID, "My Villain", "Creature — Human Villain", 2, 2)
	declareAttack(t, g, opp.ID, villain)
	passPriorityAroundTable(t, g)
	if hasEffectiveKeyword(t, g, villain, "flying") {
		t.Error("Vulture only triggers when he attacks")
	}
}

// --- Wall Crawl ---

func TestSpmB10WallCrawlMakesASpiderAndGainsOneLifePerSpider(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	b12Creature(g, me.ID, "Old Spider", "Creature — Spider", 1, 1)
	b12Creature(g, opp.ID, "Their Spider", "Creature — Spider", 1, 1)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	life := me.Life

	castCatalogSpell(t, g, "Wall Crawl", "Enchantment", spmB10WallCrawlOracle, nil)
	passPriorityAroundTable(t, g)
	if got := countBattlefieldByName(g, "Spider"); got != 1 {
		t.Fatalf("Spider tokens = %d, want 1", got)
	}
	// The new token and the old Spider: two Spiders you control.
	if me.Life != life+2 {
		t.Errorf("life %d -> %d, want +2", life, me.Life)
	}
	var token game.Card
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Spider" && c.IsToken() && c.Controller == me.ID {
			token = c
		}
	}
	if token.InstanceID == uuid.Nil {
		t.Fatal("no Spider token")
	}
	assertKeywords(t, g, token.InstanceID, "reach")
	if p := effectivePower(t, g, token.InstanceID); p != 3 {
		t.Errorf("token power %d, want 3 (2 +1)", p)
	}
	if p := effectivePower(t, g, bear); p != 2 {
		t.Errorf("a non-Spider got the bonus: power %d", p)
	}
}

func TestSpmB10WallCrawlPumpsOnlyYourSpiders(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Wall Crawl", "Enchantment", spmB10WallCrawlOracle, 0, 0)
	mine := b12Creature(g, me.ID, "Old Spider", "Creature — Spider", 1, 1)
	theirs := b12Creature(g, opp.ID, "Their Spider", "Creature — Spider", 1, 1)
	if p := effectivePower(t, g, mine); p != 2 {
		t.Errorf("my Spider power %d, want 2", p)
	}
	if p := effectivePower(t, g, theirs); p != 1 {
		t.Errorf("their Spider power %d, want 1", p)
	}
}

func TestSpmB10WallCrawlSpidersCantBeBlockedByDefenders(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Wall Crawl", "Enchantment", spmB10WallCrawlOracle, 0, 0)
	spider := brCreature(g, me.ID, "Old Spider", "Creature — Spider", 2, 2)
	bear := brCreature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	wall := brCreature(g, opp.ID, "Wall", "Creature — Wall", 0, 4, "defender")
	blocker := brCreature(g, opp.ID, "Blocker", "Creature — Bear", 2, 2)
	brAttack(t, g, spider, bear)

	brRefusal(t, g.DeclareBlocker(wall, spider), game.BlockReasonCantBeBlockedBy)
	if err := g.DeclareBlocker(wall, bear); err != nil {
		t.Errorf("a defender should still block a non-Spider: %v", err)
	}
	if err := g.DeclareBlocker(blocker, spider); err != nil {
		t.Errorf("a creature without defender should block the Spider: %v", err)
	}
}

// --- Web of Life and Destiny ---

func TestSpmB10WebHasConvoke(t *testing.T) {
	spec, ok := Lookup(spmB10WebOracle)
	if !ok || spec.TapCost == nil {
		t.Fatal("Web of Life and Destiny must carry convoke as its tap cost")
	}
}

func TestSpmB10WebPutsACreatureFromTheTopFiveOntoTheBattlefieldAtYourCombat(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	b12Push(g, me.ID, "Web of Life and Destiny", "Enchantment", spmB10WebOracle, 0, 0)
	g.WithWriteLock(func() { me.Library.Cards = nil })
	pushLibraryCardForTest(me, game.Card{Name: "F1", TypeLine: "Sorcery"})
	pushLibraryCardForTest(me, game.Card{Name: "F2", TypeLine: "Sorcery"})
	inFive := pushLibraryCardForTest(me, game.Card{Name: "In Five", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	pushLibraryCardForTest(me, game.Card{Name: "F3", TypeLine: "Sorcery"})
	pushLibraryCardForTest(me, game.Card{Name: "F4", TypeLine: "Sorcery"})
	sixth := pushLibraryCardForTest(me, game.Card{Name: "Sixth", TypeLine: "Creature — Bear", Power: 3, Toughness: 3})
	advanceToMain(t, g)
	advanceTo(t, g, game.StepBeginCombat)
	passPriorityAroundTable(t, g)

	pick := latestChooseCardsFor(g, me.ID)
	if pick == nil {
		t.Fatal("no choose_cards prompt")
	}
	if !hasID(pick.ChooseCards, inFive) || hasID(pick.ChooseCards, sixth) || len(pick.ChooseCards) != 1 {
		t.Fatalf("offered %v; want only the creature among the top five", pick.ChooseCards)
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{inFive}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(inFive) {
		t.Error("the chosen creature is not on the battlefield")
	}
	if me.Library.Size() == 0 || me.Library.Cards[len(me.Library.Cards)-1].InstanceID != sixth {
		t.Error("the sixth card should now be on top, with the rest on the bottom")
	}
	if me.Library.Size() != 5 {
		t.Errorf("library = %d, want 5", me.Library.Size())
	}
}

func TestSpmB10WebMayDeclineAndTheRestGoToTheBottom(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	b12Push(g, me.ID, "Web of Life and Destiny", "Enchantment", spmB10WebOracle, 0, 0)
	g.WithWriteLock(func() { me.Library.Cards = nil })
	first := pushLibraryCardForTest(me, game.Card{Name: "In Five", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	for i := 0; i < 4; i++ {
		pushLibraryCardForTest(me, game.Card{Name: "Filler", TypeLine: "Sorcery"})
	}
	sixth := pushLibraryCardForTest(me, game.Card{Name: "Sixth", TypeLine: "Sorcery"})
	advanceToMain(t, g)
	advanceTo(t, g, game.StepBeginCombat)
	passPriorityAroundTable(t, g)
	pick := latestChooseCardsFor(g, me.ID)
	if pick == nil {
		t.Fatal("no choose_cards prompt")
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, nil); err != nil {
		t.Fatalf("declining: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(first) || !me.Library.Contains(first) {
		t.Error("declined, but the creature moved")
	}
	if me.Library.Cards[len(me.Library.Cards)-1].InstanceID != sixth {
		t.Error("the sixth card should be on top")
	}
}

func TestSpmB10WebDoesNothingOnAnOpponentsCombat(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	b12Push(g, opp.ID, "Web of Life and Destiny", "Enchantment", spmB10WebOracle, 0, 0)
	pushLibraryCardForTest(opp, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	advanceToMain(t, g)
	advanceTo(t, g, game.StepBeginCombat)
	passPriorityAroundTable(t, g)
	if latestChooseCardsFor(g, opp.ID) != nil || latestChooseCardsFor(g, me.ID) != nil {
		t.Error("the trigger fired on another player's combat")
	}
}
