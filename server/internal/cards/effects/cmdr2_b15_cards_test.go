package effects

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Commander-deck wave 2, batch 15 (#5368): land recursion and self-mill for lands.

const (
	c2b15WreckOracle           = "7e10c141-dfb0-409b-b8c9-cdfe93b42a06"
	c2b15KirriOracle           = "f03bc648-08e4-4ecd-b339-ff72b00367c2"
	c2b15NantukoOracle         = "b8dab16c-681c-4abe-b164-fcfb5dc66678"
	c2b15ScaretillerOracle     = "e4b32e47-5484-4ee9-bac3-4ba37558a56a"
	c2b15MendingOracle         = "9b0b8f00-6367-4dc9-b33e-680470b3a36c"
	c2b15MulchOracle           = "12c03a03-36bf-487e-b8be-e55399525da5"
	c2b15EccentricFarmerOracle = "b51a852c-ccde-4d23-a3f3-a002f7eacdab"
	c2b15WindingWayOracle      = "4e48e76c-8756-4378-a722-2e2929feb74f"
	c2b15CavalierOracle        = "96b909ab-55f1-4cde-a675-2504de3da772"
)

func c2b15Forest() game.Card { return game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"} }
func c2b15Bear() game.Card {
	return game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2}
}
func c2b15Spell() game.Card { return game.Card{Name: "Opt", TypeLine: "Instant"} }

func TestC2B15MulchTakesAllLandsAndBinsTheRest(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	lib := dftMPLibraryTop(me, c2b15Forest(), c2b15Bear(), c2b15Forest(), c2b15Spell(), c2b15Bear())
	dftABCast(t, g, me, game.Card{Name: "Mulch", TypeLine: "Sorcery", OracleID: c2b15MulchOracle})
	for _, i := range []int{0, 2} {
		if !me.Hand.Contains(lib[i]) {
			t.Errorf("land %d is not in hand", i)
		}
	}
	for _, i := range []int{1, 3} {
		if !me.Graveyard.Contains(lib[i]) {
			t.Errorf("nonland %d is not in the graveyard", i)
		}
	}
	if !me.Library.Contains(lib[4]) {
		t.Error("the fifth card left the library")
	}
}

func TestC2B15WindingWayTakesTheChosenType(t *testing.T) {
	for _, tc := range []struct {
		name         string
		index        int
		inHand, bins []int
	}{
		{"creature", 0, []int{1, 3}, []int{0, 2}},
		{"land", 1, []int{0, 2}, []int{1, 3}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			advanceToMain(t, g)
			lib := dftMPLibraryTop(me, c2b15Forest(), c2b15Bear(), c2b15Forest(), c2b15Bear(), c2b15Bear())
			dftABCast(t, g, me, game.Card{Name: "Winding Way", TypeLine: "Sorcery", OracleID: c2b15WindingWayOracle})
			answerOptionPick(t, g, me.ID, tc.index)
			for _, i := range tc.inHand {
				if !me.Hand.Contains(lib[i]) {
					t.Errorf("card %d is not in hand", i)
				}
			}
			for _, i := range tc.bins {
				if !me.Graveyard.Contains(lib[i]) {
					t.Errorf("card %d is not in the graveyard", i)
				}
			}
			if !me.Library.Contains(lib[4]) {
				t.Error("the fifth card left the library")
			}
		})
	}
}

func TestC2B15EccentricFarmerMillsThreeThenMayReturnALand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	g.WithWriteLock(func() { me.Graveyard.Cards = nil })
	milled := dftMPLibraryTop(me, c2b15Forest(), c2b15Spell(), c2b15Bear())
	dftABCast(t, g, me, game.Card{Name: "Eccentric Farmer", TypeLine: "Creature — Human Peasant",
		OracleID: c2b15EccentricFarmerOracle, Power: 2, Toughness: 3})
	for _, id := range milled {
		if !me.Graveyard.Contains(id) {
			t.Fatalf("%s was not milled", id)
		}
	}
	c := chooseCardsChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no pick after the mill")
	}
	if c.ChooseMin != 0 || !slices.Equal(c.ChooseCards, []uuid.UUID{milled[0]}) {
		t.Fatalf("pick offers %v (min %d), want only the land, optionally", c.ChooseCards, c.ChooseMin)
	}
	answerChooseCards(t, g, me.ID, milled[0])
	if !me.Hand.Contains(milled[0]) {
		t.Error("the land is not in hand")
	}
}

func TestC2B15EccentricFarmerMayDecline(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	g.WithWriteLock(func() { me.Graveyard.Cards = nil })
	land := dftMPLibraryTop(me, c2b15Forest(), c2b15Spell(), c2b15Bear())[0]
	dftABCast(t, g, me, game.Card{Name: "Eccentric Farmer", TypeLine: "Creature — Human Peasant",
		OracleID: c2b15EccentricFarmerOracle, Power: 2, Toughness: 3})
	answerChooseCards(t, g, me.ID)
	if !me.Graveyard.Contains(land) || me.Hand.Contains(land) {
		t.Error("declining still returned the land")
	}
}

func TestC2B15CavalierOfThornsPutsALandAndBinsTheRest(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	lib := dftMPLibraryTop(me, c2b15Bear(), c2b15Forest(), c2b15Spell(), c2b15Forest(), c2b15Bear(), c2b15Bear())
	dftABCast(t, g, me, game.Card{Name: "Cavalier of Thorns", TypeLine: "Creature — Elemental Knight",
		OracleID: c2b15CavalierOracle, Power: 5, Toughness: 6})
	ch := chooseCardsChoiceFor(g, me.ID)
	if ch == nil {
		t.Fatal("no pick of a land")
	}
	if ch.ChooseMin != 1 || ch.ChooseMax != 1 {
		t.Errorf("want exactly one land, got min %d max %d", ch.ChooseMin, ch.ChooseMax)
	}
	answerChooseCards(t, g, me.ID, lib[1])
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(lib[1]) {
		t.Error("the chosen land is not on the battlefield")
	}
	for _, i := range []int{0, 2, 3, 4} {
		if !me.Graveyard.Contains(lib[i]) {
			t.Errorf("revealed card %d is not in the graveyard", i)
		}
	}
	if !me.Library.Contains(lib[5]) {
		t.Error("the sixth card left the library")
	}
}

// c2b15Attack attacks with a permanent so it becomes tapped.
func c2b15Attack(t *testing.T, g *game.Game, me *game.Player, id uuid.UUID) {
	t.Helper()
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(id, g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)].ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
}

func TestC2B15ScaretillerHandModePutsALandTapped(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	scare := b12Push(g, me.ID, "Scaretiller", "Artifact Creature — Scarecrow", c2b15ScaretillerOracle, 1, 4)
	land := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: land, Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
	c2b15Attack(t, g, me, scare)
	chooseModeNow(t, g, me.ID, []int{0}, 0) // no land in the graveyard: only the hand bullet
	passPriorityAroundTable(t, g)
	answerChooseCards(t, g, me.ID, land)
	var tapped bool
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == land {
			tapped = c.Tapped
		}
	}
	if !g.Battlefield.Contains(land) || !tapped {
		t.Errorf("the land is on the battlefield tapped? on=%v tapped=%v", g.Battlefield.Contains(land), tapped)
	}
}

func TestC2B15ScaretillerGraveyardModeReturnsALandTapped(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	scare := b12Push(g, me.ID, "Scaretiller", "Artifact Creature — Scarecrow", c2b15ScaretillerOracle, 1, 4)
	land := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: land, Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
	bear := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: bear, Name: "Bear", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	c2b15Attack(t, g, me, scare)
	// The hand bullet is always offered; the graveyard one has a land to take.
	chooseModeNow(t, g, me.ID, []int{0, 1}, 1)
	answerPickTargetIfOpen(t, g, land)
	passPriorityAroundTable(t, g)
	var tapped bool
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == land {
			tapped = c.Tapped
		}
	}
	if !g.Battlefield.Contains(land) || !tapped {
		t.Errorf("the land is on the battlefield tapped? on=%v tapped=%v", g.Battlefield.Contains(land), tapped)
	}
	if !me.Graveyard.Contains(bear) {
		t.Error("the creature card left the graveyard")
	}
}

func TestC2B15MendingChapterOneMillsThenMayReturnACreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	g.WithWriteLock(func() { me.Graveyard.Cards = nil })
	old := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: old, Name: "Old Bear", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	milled := dftMPLibraryTop(me, c2b15Bear(), c2b15Spell())
	dftABCast(t, g, me, game.Card{Name: "The Mending of Dominaria", TypeLine: "Enchantment — Saga", OracleID: c2b15MendingOracle})
	for _, id := range milled {
		if !me.Graveyard.Contains(id) {
			t.Fatal("a card was not milled")
		}
	}
	c := chooseCardsChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no creature pick after the mill")
	}
	if c.ChooseMin != 0 || !slices.Contains(c.ChooseCards, old) || !slices.Contains(c.ChooseCards, milled[0]) || len(c.ChooseCards) != 2 {
		t.Fatalf("pick offers %v (min %d), want the two creature cards, optionally", c.ChooseCards, c.ChooseMin)
	}
	answerChooseCards(t, g, me.ID, old)
	if !me.Hand.Contains(old) {
		t.Error("the creature is not in hand")
	}
}

func TestC2B15MendingChapterThreeReturnsLandsThenShufflesTheGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	saga := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "The Mending of Dominaria", TypeLine: "Enchantment — Saga", OracleID: c2b15MendingOracle,
		Owner: me.ID, Controller: me.ID, Counters: map[string]int{game.CounterLore: 2},
	})
	g.WithWriteLock(func() { me.Graveyard.Cards = nil })
	var lands []uuid.UUID
	for i := 0; i < 2; i++ {
		id := uuid.New()
		lands = append(lands, id)
		me.Graveyard.PushTop(game.Card{InstanceID: id, Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
	}
	bear := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: bear, Name: "Bear", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	advanceToPrecombatMainOf(t, g, g.Turn.ActiveSeat)
	passPriorityAroundTable(t, g)
	if got := loreCountersOn(g, saga); got != -1 && got != 3 {
		t.Errorf("lore counters %d, want 3 or sacrificed", got)
	}
	for _, id := range lands {
		if !g.Battlefield.Contains(id) {
			t.Errorf("land %s did not return to the battlefield", id)
		}
	}
	if me.Graveyard.Contains(bear) || !me.Library.Contains(bear) {
		t.Error("the graveyard was not shuffled into the library")
	}
}

func TestC2B15KirriPumpsOtherPlantsAndTreefolkAndRecursOnPostcombatMain(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	kirri := b12Push(g, me.ID, "Kirri, Talented Sprout", "Legendary Creature — Plant Druid", c2b15KirriOracle, 0, 3)
	plant := b12Push(g, me.ID, "Sprout", "Creature — Plant", "", 1, 1)
	tree := b12Push(g, me.ID, "Ent", "Creature — Treefolk", "", 4, 4)
	bear := b12Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	theirs := b12Push(g, g.Seats[(g.Turn.ActiveSeat+1)%4].ID, "Their Sprout", "Creature — Plant", "", 1, 1)
	for id, want := range map[uuid.UUID]int{kirri: 0, plant: 3, tree: 6, bear: 2, theirs: 1} {
		if got := ctrlEffectivePower(t, g, id); got != want {
			t.Errorf("power of %s = %d, want %d", id, got, want)
		}
	}
	g.WithWriteLock(func() { me.Graveyard.Cards = nil })
	mk := func(name, tl string) uuid.UUID {
		id := uuid.New()
		me.Graveyard.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: tl, Owner: me.ID, Controller: me.ID})
		return id
	}
	land := mk("Forest", "Basic Land — Forest")
	deadBear := mk("Dead Bear", "Creature — Bear")
	advanceTo(t, g, game.StepPostcombatMain)
	ch := pickTargetPrompt(t, g)
	if slices.Contains(ch.PickTargetCards, deadBear) {
		t.Error("a nonland non-Plant card is a legal target")
	}
	answerPickTarget(t, g, land)
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(land) {
		t.Error("the land card did not return to hand")
	}
	if !me.Graveyard.Contains(deadBear) {
		t.Error("the Bear card left the graveyard")
	}
}

func TestC2B15NantukoCultivatorDiscardsLandsForCountersAndCards(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	mk := func(name, tl string) uuid.UUID {
		id := uuid.New()
		me.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: tl, Owner: me.ID, Controller: me.ID})
		return id
	}
	f1, f2 := mk("Forest", "Basic Land — Forest"), mk("Forest", "Basic Land — Forest")
	spell := mk("Opt", "Instant")
	id := dftABCast(t, g, me, game.Card{Name: "Nantuko Cultivator", TypeLine: "Creature — Insect Druid",
		OracleID: c2b15NantukoOracle, Power: 2, Toughness: 2})
	ch := chooseCardsChoiceFor(g, me.ID)
	if ch == nil {
		t.Fatal("no discard pick")
	}
	if ch.ChooseMin != 0 || ch.ChooseMax != 2 || slices.Contains(ch.ChooseCards, spell) {
		t.Fatalf("pick offers %v (%d..%d), want the two lands, any number", ch.ChooseCards, ch.ChooseMin, ch.ChooseMax)
	}
	answerChooseCards(t, g, me.ID, f1, f2)
	passPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(f1) || !me.Graveyard.Contains(f2) {
		t.Error("the lands were not discarded")
	}
	c, _ := battlefieldCard(g, id)
	if c.Counters[game.CounterPlusOne] != 2 {
		t.Errorf("+1/+1 counters %d, want 2", c.Counters[game.CounterPlusOne])
	}
	if got := me.Hand.Size(); got != 3 {
		t.Errorf("hand %d, want 3 (the spell and two draws)", got)
	}
}

func TestC2B15NantukoCultivatorMayDiscardNone(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	me.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
	id := dftABCast(t, g, me, game.Card{Name: "Nantuko Cultivator", TypeLine: "Creature — Insect Druid",
		OracleID: c2b15NantukoOracle, Power: 2, Toughness: 2})
	answerChooseCards(t, g, me.ID)
	passPriorityAroundTable(t, g)
	c, _ := battlefieldCard(g, id)
	if c.Counters[game.CounterPlusOne] != 0 || me.Hand.Size() != 1 {
		t.Errorf("declined but counters=%d hand=%d", c.Counters[game.CounterPlusOne], me.Hand.Size())
	}
}

func TestC2B15WreckAndRebuildDestroysAnArtifactOrEnchantment(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	rock := b12Push(g, opp.ID, "Rock", "Artifact", "", 0, 0)
	ench := b12Push(g, opp.ID, "Aura", "Enchantment", "", 0, 0)
	bear := b12Push(g, opp.ID, "Bear", "Creature — Bear", "", 2, 2)
	if err := castCatalogSpellErr(t, g, "Wreck and Rebuild", "Sorcery", c2b15WreckOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}}); err == nil {
		t.Error("a creature was accepted as the target of the destroy mode")
	}
	castModal(t, g, "Wreck and Rebuild", "Sorcery", c2b15WreckOracle, []int{0}, dftABCardRef(ench))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(ench) || !g.Battlefield.Contains(rock) {
		t.Error("the enchantment should be destroyed and the artifact left alone")
	}
}

func TestC2B15WreckAndRebuildMillsFiveThenPutsALandTappedAndFlashesBack(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	g.WithWriteLock(func() { me.Graveyard.Cards = nil })
	milled := dftMPLibraryTop(me, c2b15Bear(), c2b15Forest(), c2b15Spell(), c2b15Bear(), c2b15Bear())
	id := castModal(t, g, "Wreck and Rebuild", "Sorcery", c2b15WreckOracle, []int{1}, nil)
	passPriorityAroundTable(t, g)
	for _, m := range milled {
		if !me.Graveyard.Contains(m) {
			t.Fatal("a card was not milled")
		}
	}
	ch := chooseCardsChoiceFor(g, me.ID)
	if ch == nil || ch.ChooseMin != 0 || !slices.Equal(ch.ChooseCards, []uuid.UUID{milled[1]}) {
		t.Fatalf("want an optional pick of only the land, got %+v", ch)
	}
	answerChooseCards(t, g, me.ID, milled[1])
	var tapped bool
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == milled[1] {
			tapped = c.Tapped
		}
	}
	if !g.Battlefield.Contains(milled[1]) || !tapped {
		t.Errorf("land on the battlefield tapped? on=%v tapped=%v", g.Battlefield.Contains(milled[1]), tapped)
	}
	if !me.Graveyard.Contains(id) {
		t.Fatal("Wreck and Rebuild is not in the graveyard")
	}
	// Flashback {3}{R}{G}: the other bullet works from the graveyard and exiles it.
	rock := b12Push(g, g.Seats[(g.Turn.ActiveSeat+1)%4].ID, "Rock", "Artifact", "", 0, 0)
	fillPoolColored(me, "R", 1)
	fillPoolColored(me, "G", 1)
	fillPoolColored(me, "C", 3)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{FromZone: "graveyard", AlternativeCost: "flashback",
		Modes: []int{0}, Targets: dftABCardRef(rock)}); err != nil {
		t.Fatalf("flashback cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("the flashed-back mode did not destroy the artifact")
	}
	if !g.Exile.Contains(id) {
		t.Error("the flashed-back spell is not in exile")
	}
}
