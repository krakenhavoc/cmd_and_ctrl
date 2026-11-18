package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_rpr_b26_cards_test.go — reprint pass, batch 26 (#5908). Each test plays
// the real catalog entry through the engine and asserts what the printed text
// says.

const (
	rprB26RaidersKarve = "1e9b3c9e-07ae-4de5-8607-93773825f655"
	rprB26Riddlesmith  = "e3b9c9e3-5a35-4dd2-ab17-d1b6a140b185"
	rprB26ShiftingShad = "564b1253-2e1f-46c5-b5ae-3932127ca614"
	rprB26Spearbreaker = "025f8db8-11d1-4c59-8e1f-004c9c843fd8"
	rprB26Ulasht       = "796f1e1f-7fec-429d-82ae-125f541d6cc7"
	rprB26Vastwood     = "8ba2a7ad-a09b-4f52-9c76-c70510dc4feb"
	rprB26Engineer     = "b36accb4-cde6-4b1d-b549-8c975cfa999d"
	rprB26Weatherlight = "73b3682f-396a-4ebe-aaac-6ce3ee36283a"
	rprB26Dissenter    = "11cb509d-21af-48f1-b355-135ebd3e4bd1"
	rprB26Craving      = "78725353-9274-420a-b722-add0f43c444e"
	rprB26AngelFlight  = "9b6a74fb-c48f-4d85-b4b3-96eb80b9efc7"
)

func TestRprB26EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		rprB26RaidersKarve: "Raiders' Karve",
		rprB26Riddlesmith:  "Riddlesmith",
		rprB26ShiftingShad: "Shifting Shadow",
		rprB26Spearbreaker: "Spearbreaker Behemoth",
		rprB26Ulasht:       "Ulasht, the Hate Seed",
		rprB26Vastwood:     "Vastwood Surge",
		rprB26Engineer:     "Vedalken Engineer",
		rprB26Weatherlight: "Weatherlight",
		rprB26Dissenter:    "Doomed Dissenter",
		rprB26Craving:      "Ancient Craving",
		rprB26AngelFlight:  "Angel of Flight Alabaster",
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

func TestRprB26AncientCravingDrawsThreeAndLosesThree(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedLibrary(me, "A", "B", "C", "D")
	life, hand := me.Life, me.Hand.Size()
	castCatalogSpell(t, g, "Ancient Craving", "Sorcery", rprB26Craving, nil)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 3 {
		t.Errorf("hand grew by %d (the spell left it), want 3 net of the cast card", got+1)
	}
	if me.Life != life-3 {
		t.Errorf("life %d, want %d", me.Life, life-3)
	}
}

func TestRprB26DoomedDissenterLeavesAZombieWhenItDies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	d := pushCatalogPermanent(g, me.ID, "Doomed Dissenter", "Creature — Human", rprB26Dissenter, false)
	p1Destroy(t, g, d)
	passPriorityAroundTable(t, g)
	zombies := socB17Names(g, me.ID, "Zombie")
	if len(zombies) != 1 {
		t.Fatalf("%d Zombies, want 1", len(zombies))
	}
	if zombies[0].Power != 2 || zombies[0].Toughness != 2 {
		t.Errorf("Zombie is %d/%d, want 2/2", zombies[0].Power, zombies[0].Toughness)
	}
}

func TestRprB26RiddlesmithLootsOnArtifactSpellsOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	msc18Library(me, 6)
	pushCatalogPermanent(g, me.ID, "Riddlesmith", "Creature — Human Artificer", rprB26Riddlesmith, false)

	// A creature spell does nothing.
	castCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	if latestChoiceOfKind(g, game.PendingChoiceConfirm) != nil {
		t.Fatal("a creature spell asked for a loot")
	}
	passPriorityAroundTable(t, g)

	// An artifact spell: decline, nothing; accept, loot.
	castCatalogSpell(t, g, "Trinket", "Artifact", "", nil)
	passPriorityAroundTable(t, g)
	ask := latestChoiceOfKind(g, game.PendingChoiceConfirm)
	if ask == nil || ask.Chooser != me.ID {
		t.Fatal("no \"draw a card?\" prompt for an artifact spell")
	}
	hand := me.Hand.Size()
	if err := g.ResolveConfirm(ask.ID, me.ID, true); err != nil {
		t.Fatalf("ResolveConfirm: %v", err)
	}
	if me.Hand.Size() != hand+1 {
		t.Fatalf("hand %d after drawing, want %d", me.Hand.Size(), hand+1)
	}
	discardFromHand(t, g, me.ID)
	if me.Hand.Size() != hand {
		t.Errorf("hand %d after the discard, want %d", me.Hand.Size(), hand)
	}
}

func TestRprB26SpearbreakerGrantsIndestructibleOnlyToPowerFiveOrMore(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	beh := pushCatalogPermanent(g, me.ID, "Spearbreaker Behemoth", "Creature — Beast", rprB26Spearbreaker, false)
	big := b12Push(g, me.ID, "Big", "Creature — Beast", "", 5, 5)
	small := b12Push(g, me.ID, "Small", "Creature — Bear", "", 2, 2)
	floatForTest(g, me, "GG")
	err := g.ActivateCatalogAbility(me.ID, beh, 0, game.ActivateAbilityParams{Strict: true, Targets: cardTarget(small)})
	if err == nil {
		t.Fatal("a 2/2 was accepted as the target")
	}
	if err := g.ActivateCatalogAbility(me.ID, beh, 0, game.ActivateAbilityParams{Strict: true, Targets: cardTarget(big)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !rprB26Has(t, g, big, "indestructible") {
		t.Error("the 5/5 did not gain indestructible")
	}
	if !rprB26Has(t, g, beh, "indestructible") {
		t.Error("Behemoth is not indestructible itself")
	}
}

func TestRprB26VastwoodSurgeFetchesTwoTappedAndKickedPutsCounters(t *testing.T) {
	for _, tc := range []struct {
		name     string
		optional []int
		counters int
	}{
		{"unkicked", nil, 0},
		{"kicked", []int{0}, 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[0]
			ids := seedBasicsInLibrary(me, 2)
			bear := seedBear(g, me.ID)
			if _, err := castWithOptionalCosts(t, g, "Vastwood Surge", "Sorcery", rprB26Vastwood, nil, tc.optional, nil); err != nil {
				t.Fatalf("CastSpell: %v", err)
			}
			passPriorityAroundTable(t, g)
			answerSearchByID(t, g, me.ID, ids...)
			for _, id := range ids {
				if !g.Battlefield.Contains(id) {
					t.Fatal("a basic land did not reach the battlefield")
				}
				if c, _ := g.LookupCardForEffect(id); !c.Tapped {
					t.Error("a fetched land entered untapped")
				}
			}
			if got := counterOn(g, bear, game.CounterPlusOne); got != tc.counters {
				t.Errorf("Bear has %d +1/+1 counters, want %d", got, tc.counters)
			}
		})
	}
}

func TestRprB26VedalkenEngineerAddsTwoOfOneColourForArtifactsOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	eng := b12Push(g, me.ID, "Vedalken Engineer", "Creature — Vedalken Artificer", rprB26Engineer, 1, 1)
	advanceToMain(t, g)
	untapForTest(g, eng)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == eng {
				g.Battlefield.Cards[i].SummonedThisTurn = false
			}
		}
	})
	if err := g.ActivateManaAbility(me.ID, eng, 0, game.ManaAbilityParams{Colors: []string{"U"}}); err != nil {
		t.Fatalf("tap the Engineer: %v", err)
	}
	cost, _ := game.ParseCost("{2}")
	two, _ := game.ParseCost("{U}{U}")
	rock := game.Card{Name: "Rock", TypeLine: "Artifact", ManaCost: "{2}"}
	bear := game.Card{Name: "Bear", TypeLine: "Creature — Bear", ManaCost: "{2}"}
	if !me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForCast(rock)) || !me.ManaPool.CanPayFor(two, 0, game.ManaSpendForAbility(rock)) {
		t.Error("the two mana refused an artifact spell or an artifact's ability")
	}
	if me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForCast(bear)) || me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForAbility(bear)) {
		t.Error("the mana paid for a nonartifact spell or ability")
	}
}

func TestRprB26VehiclesCrewThree(t *testing.T) {
	for _, tc := range []struct{ name, oracle string }{
		{"Raiders' Karve", rprB26RaidersKarve},
		{"Weatherlight", rprB26Weatherlight},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[0]
			veh := b12Push(g, me.ID, tc.name, "Artifact — Vehicle", tc.oracle, 4, 4)
			small := b12Creature(g, me.ID, "Squire", "Creature — Human", 2, 2)
			advanceToMain(t, g)
			if c, _ := battlefieldCard(g, veh); c.IsCreature() {
				t.Fatal("a Vehicle is a creature before it is crewed")
			}
			if err := g.ActivateCatalogAbility(me.ID, veh, 0, game.ActivateAbilityParams{Strict: true, CrewIDs: []uuid.UUID{small}}); err == nil {
				t.Error("power 2 crewed a Crew 3 Vehicle")
			}
			big := b12Creature(g, me.ID, "Giant", "Creature — Giant", 3, 3)
			crewBy(t, g, me.ID, veh, big)
			passPriorityAroundTable(t, g)
			if v := s58p6View(t, g, veh); !v.IsCreature() {
				t.Error("power 3 did not crew the Vehicle")
			}
		})
	}
}

func TestRprB26AngelOfFlightAlabasterReturnsASpiritAtUpkeep(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Angel of Flight Alabaster", "Creature — Angel", rprB26AngelFlight, false)
	spirit := pushGraveyardPermanent(me, "Ghost", "Creature — Spirit", "{1}{W}")
	bear := pushGraveyardPermanent(me, "Grizzly", "Creature — Bear", "{1}{G}")
	advanceToUpkeepOfSeat(t, g, 0)
	if latestPickTarget(g, me.ID) == nil {
		t.Fatal("no target prompt for the upkeep trigger")
	}
	if pick := latestPickTarget(g, me.ID); g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: bear}) == nil {
		t.Fatal("a non-Spirit card was accepted as the target")
	}
	pickTriggerTarget(t, g, me.ID, spirit)
	passPriorityAroundTable(t, g)
	inHand := func(id uuid.UUID) bool {
		for _, c := range me.Hand.Cards {
			if c.InstanceID == id {
				return true
			}
		}
		return false
	}
	if !inHand(spirit) {
		t.Error("the Spirit card did not return to hand")
	}
	if inHand(bear) {
		t.Error("a non-Spirit card returned")
	}
}

func TestRprB26UlashtEntersWithACounterPerOtherRedAndGreenCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Red", TypeLine: "Creature — Goblin",
		Colors: []string{"R"}, Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
	pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Green", TypeLine: "Creature — Elf",
		Colors: []string{"G"}, Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
	pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Both", TypeLine: "Creature — Hybrid",
		Colors: []string{"R", "G"}, Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
	pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Blue", TypeLine: "Creature — Merfolk",
		Colors: []string{"U"}, Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
	id, err := msc17Cast(t, g, me, game.Card{
		Name: "Ulasht, the Hate Seed", TypeLine: "Legendary Creature — Hellion Hydra", OracleID: rprB26Ulasht,
		Colors: []string{"R", "G"},
	}, game.CastSpellParams{})
	if err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := counterOn(g, id, game.CounterPlusOne); got != 4 {
		t.Fatalf("Ulasht has %d counters, want 4 (red 2, green 2, itself not counted)", got)
	}

	// Remove a counter for a Saproling.
	floatForTest(g, me, "R")
	err = g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true, Modes: []int{1}})
	if err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := counterOn(g, id, game.CounterPlusOne); got != 3 {
		t.Errorf("Ulasht has %d counters after paying one, want 3", got)
	}
	if len(socB17Names(g, me.ID, "Saproling")) != 1 {
		t.Error("no Saproling token")
	}
}

func TestRprB26RaidersKarveMayPutATopLandOntoTheBattlefieldTapped(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	karve := pushCatalogPermanent(g, me.ID, "Raiders' Karve", "Artifact Creature — Vehicle", rprB26RaidersKarve, false)
	land := seedLibraryTop(me, "Forest", "Basic Land — Forest", "")
	declareAttack(t, g, opp.ID, karve)
	passPriorityAroundTable(t, g)
	answerChooseCards(t, g, me.ID, land)
	if !g.Battlefield.Contains(land) {
		t.Fatal("the land did not enter")
	}
	if c, _ := g.LookupCardForEffect(land); !c.Tapped {
		t.Error("the land entered untapped")
	}
}

func TestRprB26RaidersKarveLeavesANonlandOnTop(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	karve := pushCatalogPermanent(g, me.ID, "Raiders' Karve", "Artifact Creature — Vehicle", rprB26RaidersKarve, false)
	top := seedLibraryTop(me, "Bolt", "Instant", "{R}")
	declareAttack(t, g, opp.ID, karve)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(top) || me.Library.Cards[len(me.Library.Cards)-1].InstanceID != top {
		t.Error("the nonland card did not stay on top of the library")
	}
}

func TestRprB26WeatherlightTakesAHistoricCardFromTheTopFive(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	w := pushCatalogPermanent(g, me.ID, "Weatherlight", "Legendary Artifact Creature — Vehicle", rprB26Weatherlight, false)
	rest := seedLibrary(me, "A", "B", "C", "D", "E", "F")
	rock := seedLibraryTop(me, "Rock", "Artifact", "{2}")
	declareAttack(t, g, opp.ID, w)
	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)
	answerChooseCards(t, g, me.ID, rock)
	inHand := false
	for _, c := range me.Hand.Cards {
		if c.InstanceID == rock {
			inHand = true
		}
	}
	if !inHand {
		t.Fatal("the historic card was not put into hand")
	}
	_ = rest
}

func TestRprB26ShiftingShadowGivesHasteAndReplacesTheCreatureAtUpkeep(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	host := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Host", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	aura := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Shifting Shadow",
		TypeLine: "Enchantment — Aura", OracleID: rprB26ShiftingShad, Owner: me.ID, Controller: me.ID})
	g.WithWriteLock(func() {
		if err := g.AttachForEffect(aura, game.TargetRef{Kind: game.TargetCard, ID: host}); err != nil {
			t.Fatalf("attach: %v", err)
		}
	})
	if !rprB26Has(t, g, host, "haste") {
		t.Error("the enchanted creature has no haste")
	}
	seedLibrary(me, "Spell One")
	creature := pushLibraryCardForTest(me, game.Card{Name: "Target Creature", TypeLine: "Creature — Wurm", Power: 4, Toughness: 4})
	// Library order top-first: Spell One, Target Creature; seedLibrary pushed on top.
	advanceToUpkeepOfSeat(t, g, 0)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(host) {
		t.Error("the enchanted creature survived")
	}
	if !g.Battlefield.Contains(creature) {
		t.Fatal("the revealed creature card did not enter")
	}
	var attachedTo uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == aura {
			attachedTo = c.AttachedTo.ID
		}
	}
	if attachedTo != creature {
		t.Errorf("Shifting Shadow is attached to %v, want the new creature %v", attachedTo, creature)
	}
}

// rprB26Has reads a keyword through the layered view.
func rprB26Has(t *testing.T, g *game.Game, id uuid.UUID, kw string) bool {
	t.Helper()
	c := s58p6View(t, g, id)
	return game.HasKeyword(&c, kw)
}
