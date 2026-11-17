package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_one_b10_test.go — Phyrexia: All Will Be One, batch 10 (#5479): the
// oil cards. Each test plays the real catalog entry through the engine.

const (
	s82b10Ambassador  = "3088ee75-3b51-4cbc-8987-76cb93f416b1"
	s82b10Font        = "ec1fcec2-ba7a-44d7-b251-22a7cd417493"
	s82b10Steward     = "5e3d2b03-4c44-49a3-ab11-871ca9c372f7"
	s82b10Raptor      = "6096b221-7b6e-4528-8b5d-8ad0fcef3fba"
	s82b10Cultivator  = "7a49731b-bd95-4f0a-ac47-1aa318c12fec"
	s82b10Sac         = "478cc55b-dd71-4adc-909d-c39d435587ee"
	s82b10Blisterzoa  = "ca130ba1-a4a8-4225-9799-75d56bac6fb3"
	s82b10Strider     = "85488433-3c23-4a29-b063-53dba71f26c4"
	s82b10Seer        = "bcbb318d-0ebc-48a2-afd7-69cf5351893d"
	s82b10Meldweb     = "b66dd751-fef7-4310-9c09-8d1b1fc78fef"
	s82b10Engraver    = "a7ee67a6-0ac7-4079-a947-4e52d09ace8e"
	s82b10Immobilizer = "c2c7a9c2-28f8-4cf0-a042-420f6a4ba8bd"
	s82b10Mantis      = "6755a7af-5b55-4245-a777-27e69979479b"
)

// s82b10Push puts a catalog permanent onto the battlefield, ready to act,
// carrying `oil` oil counters.
func s82b10Push(g *game.Game, owner uuid.UUID, name, typeLine, oracle string, p, tough, oil int) uuid.UUID {
	c := game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: p, Toughness: tough, Owner: owner, Controller: owner,
	}
	if oil > 0 {
		c.Counters = map[string]int{oilCounter: oil}
	}
	return pushBattlefieldCardWithTimestamp(g, c)
}

func s82b10Activate(t *testing.T, g *game.Game, who, card uuid.UUID, idx int, params game.ActivateAbilityParams) error {
	t.Helper()
	params.Strict = true
	return g.ActivateCatalogAbility(who, card, idx, params)
}

func TestS82B10EveryCardEntersWithItsOilCounters(t *testing.T) {
	cases := []struct {
		name, typeLine, oracle string
		p, tough, oil          int
	}{
		{"Bladed Ambassador", "Creature — Phyrexian Soldier", s82b10Ambassador, 3, 1, 1},
		{"Axiom Engraver", "Creature — Phyrexian Wizard", s82b10Engraver, 1, 3, 2},
		{"Furnace Strider", "Creature — Phyrexian Beast", s82b10Strider, 4, 5, 2},
		{"Gitaxian Raptor", "Creature — Phyrexian Bird", s82b10Raptor, 1, 4, 3},
		{"Glistener Seer", "Creature — Phyrexian Advisor", s82b10Seer, 0, 3, 3},
		{"Predation Steward", "Creature — Phyrexian Elf Warrior", s82b10Steward, 2, 2, 2},
		{"Rustvine Cultivator", "Creature — Phyrexian Elf Druid", s82b10Cultivator, 1, 2, 0},
		{"Lattice-Blade Mantis", "Creature — Phyrexian Insect", s82b10Mantis, 4, 3, 2},
		{"Watchful Blisterzoa", "Creature — Phyrexian Jellyfish", s82b10Blisterzoa, 4, 4, 1},
		{"Tamiyo's Immobilizer", "Artifact", s82b10Immobilizer, 0, 0, 4},
		{"Incubation Sac", "Artifact", s82b10Sac, 0, 0, 3},
		{"Font of Progress", "Artifact", s82b10Font, 0, 0, 2},
		{"Meldweb Strider", "Artifact — Vehicle", s82b10Meldweb, 5, 5, 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			id := rfdCast(t, g, tc.name, tc.typeLine, tc.oracle, tc.p, tc.tough)
			if got := countersOn(g, id, oilCounter); got != tc.oil {
				t.Errorf("entered with %d oil counters, want %d", got, tc.oil)
			}
		})
	}
}

func TestS82B10BladedAmbassadorGainsIndestructibleForACounter(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	amb := s82b10Push(g, me.ID, "Bladed Ambassador", "Creature — Phyrexian Soldier", s82b10Ambassador, 3, 1, 1)
	if err := s82b10Activate(t, g, me.ID, amb, 0, game.ActivateAbilityParams{}); err == nil {
		t.Fatal("activated without paying {1}")
	}
	addPool(t, g, me.ID, "{C}")
	if err := s82b10Activate(t, g, me.ID, amb, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !hasEffectiveKeyword(t, g, amb, "indestructible") {
		t.Error("no indestructible")
	}
	if got := countersOn(g, amb, oilCounter); got != 0 {
		t.Errorf("%d oil counters left, want 0", got)
	}
	addPool(t, g, me.ID, "{C}")
	if err := s82b10Activate(t, g, me.ID, amb, 0, game.ActivateAbilityParams{}); err == nil {
		t.Error("activated with no oil counter left")
	}
}

func TestS82B10AxiomEngraverDiscardsToDraw(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	eng := s82b10Push(g, me.ID, "Axiom Engraver", "Creature — Phyrexian Wizard", s82b10Engraver, 1, 3, 2)
	if err := s82b10Activate(t, g, me.ID, eng, 0, game.ActivateAbilityParams{}); err == nil {
		t.Fatal("activated without discarding a card")
	}
	pitched := pushHandCard(g, me)
	hand := me.Hand.Size()
	if err := s82b10Activate(t, g, me.ID, eng, 0, game.ActivateAbilityParams{DiscardIDs: []uuid.UUID{pitched}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Hand.Contains(pitched) || me.Hand.Size() != hand {
		t.Errorf("hand %d (had %d), pitched card still in hand: %v", me.Hand.Size(), hand, me.Hand.Contains(pitched))
	}
	if got := countersOn(g, eng, oilCounter); got != 1 {
		t.Errorf("%d oil counters left, want 1", got)
	}
	if tapped, _ := battlefieldCardTapped(g, eng); !tapped {
		t.Error("the Engraver did not tap")
	}
}

func TestS82B10FurnaceStriderGivesHasteToYourCreatureOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	str := s82b10Push(g, me.ID, "Furnace Strider", "Creature — Phyrexian Beast", s82b10Strider, 4, 5, 2)
	theirs := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	if err := s82b10Activate(t, g, me.ID, str, 0, game.ActivateAbilityParams{Targets: cardRefs(theirs)}); err == nil {
		t.Fatal("targeted an opponent's creature")
	}
	mine := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if err := s82b10Activate(t, g, me.ID, str, 0, game.ActivateAbilityParams{Targets: cardRefs(mine)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !hasEffectiveKeyword(t, g, mine, "haste") {
		t.Error("no haste")
	}
	if got := countersOn(g, str, oilCounter); got != 1 {
		t.Errorf("%d oil counters left, want 1", got)
	}
}

func TestS82B10GitaxianRaptorPumpsOncePerCounter(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	r := s82b10Push(g, me.ID, "Gitaxian Raptor", "Creature — Phyrexian Bird", s82b10Raptor, 1, 4, 3)
	for i := 0; i < 3; i++ {
		if err := s82b10Activate(t, g, me.ID, r, 0, game.ActivateAbilityParams{}); err != nil {
			t.Fatalf("activate %d: %v", i, err)
		}
		passPriorityAroundTable(t, g)
	}
	if p, tt := ptOf(t, g, r); p != 4 || tt != 1 {
		t.Errorf("Raptor is %d/%d, want 4/1", p, tt)
	}
	if err := s82b10Activate(t, g, me.ID, r, 0, game.ActivateAbilityParams{}); err == nil {
		t.Error("a fourth activation with no counters left")
	}
}

func TestS82B10GlistenerSeerTapsAndScries(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seer := s82b10Push(g, me.ID, "Glistener Seer", "Creature — Phyrexian Advisor", s82b10Seer, 0, 3, 3)
	if err := s82b10Activate(t, g, me.ID, seer, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if got := countersOn(g, seer, oilCounter); got != 2 {
		t.Errorf("%d oil counters left, want 2", got)
	}
	if tapped, _ := battlefieldCardTapped(g, seer); !tapped {
		t.Error("the Seer did not tap")
	}
	if err := s82b10Activate(t, g, me.ID, seer, 0, game.ActivateAbilityParams{}); err == nil {
		t.Error("activated while tapped")
	}
}

func TestS82B10PredationStewardPumpsAsASorcery(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	st := s82b10Push(g, me.ID, "Predation Steward", "Creature — Phyrexian Elf Warrior", s82b10Steward, 2, 2, 2)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	toMainForCost(t, g)
	addPool(t, g, me.ID, "{C}{C}{G}")
	if err := s82b10Activate(t, g, me.ID, st, 0, game.ActivateAbilityParams{Targets: cardRefs(bear)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p, tt := ptOf(t, g, bear); p != 4 || tt != 4 {
		t.Errorf("target is %d/%d, want 4/4", p, tt)
	}
	if got := countersOn(g, st, oilCounter); got != 1 {
		t.Errorf("%d oil counters left, want 1", got)
	}
}

func TestS82B10PredationStewardRefusesOutsideAMainPhase(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	st := s82b10Push(g, me.ID, "Predation Steward", "Creature — Phyrexian Elf Warrior", s82b10Steward, 2, 2, 2)
	advanceTo(t, g, game.StepBeginCombat)
	addPool(t, g, me.ID, "{C}{C}{G}")
	if err := s82b10Activate(t, g, me.ID, st, 0, game.ActivateAbilityParams{Targets: cardRefs(st)}); err == nil {
		t.Error("activated outside a main phase")
	}
}

func TestS82B10RustvineCultivatorMakesOilThenUntapsALand(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	cult := s82b10Push(g, me.ID, "Rustvine Cultivator", "Creature — Phyrexian Elf Druid", s82b10Cultivator, 1, 2, 0)
	land := b12Permanent(g, opp.ID, "Their Land", "Land")
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == land {
				g.Battlefield.Cards[i].Tapped = true
			}
		}
	})
	if err := s82b10Activate(t, g, me.ID, cult, 1, game.ActivateAbilityParams{Targets: cardRefs(land)}); err == nil {
		t.Fatal("untap ability paid with no oil counter")
	}
	if err := s82b10Activate(t, g, me.ID, cult, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("make oil: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := countersOn(g, cult, oilCounter); got != 1 {
		t.Fatalf("%d oil counters, want 1", got)
	}
	// Untap it (a new turn would do this) and spend the counter.
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == cult {
				g.Battlefield.Cards[i].Tapped = false
			}
		}
	})
	if err := s82b10Activate(t, g, me.ID, cult, 1, game.ActivateAbilityParams{Targets: cardRefs(land)}); err != nil {
		t.Fatalf("untap land: %v", err)
	}
	passPriorityAroundTable(t, g)
	if tapped, _ := battlefieldCardTapped(g, land); tapped {
		t.Error("the land is still tapped")
	}
	if got := countersOn(g, cult, oilCounter); got != 0 {
		t.Errorf("%d oil counters left, want 0", got)
	}
}

func TestS82B10LatticeBladeMantisMayTradeACounterToUntapAndGrow(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	m := s82b10Push(g, me.ID, "Lattice-Blade Mantis", "Creature — Phyrexian Insect", s82b10Mantis, 4, 3, 2)
	declareAttack(t, g, g.Seats[1].ID, m)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if got := countersOn(g, m, oilCounter); got != 1 {
		t.Errorf("%d oil counters left, want 1", got)
	}
	if p, tt := ptOf(t, g, m); p != 5 || tt != 4 {
		t.Errorf("Mantis is %d/%d, want 5/4", p, tt)
	}
	if tapped, _ := battlefieldCardTapped(g, m); tapped {
		t.Error("the Mantis is still tapped")
	}
}

func TestS82B10LatticeBladeMantisDecliningKeepsTheCounter(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	m := s82b10Push(g, me.ID, "Lattice-Blade Mantis", "Creature — Phyrexian Insect", s82b10Mantis, 4, 3, 2)
	declareAttack(t, g, g.Seats[1].ID, m)
	answerLatestTriggerPrompt(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if got := countersOn(g, m, oilCounter); got != 2 {
		t.Errorf("%d oil counters, want 2", got)
	}
	if p, tt := ptOf(t, g, m); p != 4 || tt != 3 {
		t.Errorf("Mantis is %d/%d, want 4/3", p, tt)
	}
	if tapped, _ := battlefieldCardTapped(g, m); !tapped {
		t.Error("the Mantis untapped without paying")
	}
}

func TestS82B10WatchfulBlisterzoaDrawsForItsOilWhenItDies(t *testing.T) {
	for _, oil := range []int{0, 1, 3} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		z := s82b10Push(g, me.ID, "Watchful Blisterzoa", "Creature — Phyrexian Jellyfish", s82b10Blisterzoa, 4, 4, oil)
		hand := me.Hand.Size()
		g.WithWriteLock(func() { g.DestroyPermanentsForEffect([]uuid.UUID{z}) })
		passPriorityAroundTable(t, g)
		if got := me.Hand.Size() - hand; got != oil {
			t.Errorf("with %d oil counters drew %d cards", oil, got)
		}
	}
}

func TestS82B10TamiyosImmobilizerTapsAnArtifactOrCreatureOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	imm := s82b10Push(g, me.ID, "Tamiyo's Immobilizer", "Artifact", s82b10Immobilizer, 0, 0, 4)
	land := b12Permanent(g, opp.ID, "Their Land", "Land")
	if err := s82b10Activate(t, g, me.ID, imm, 0, game.ActivateAbilityParams{Targets: cardRefs(land)}); err == nil {
		t.Fatal("targeted a land")
	}
	rock := b12Permanent(g, opp.ID, "Rock", "Artifact")
	if err := s82b10Activate(t, g, me.ID, imm, 0, game.ActivateAbilityParams{Targets: cardRefs(rock)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if tapped, _ := battlefieldCardTapped(g, rock); !tapped {
		t.Error("the artifact did not tap")
	}
	if got := countersOn(g, imm, oilCounter); got != 3 {
		t.Errorf("%d oil counters left, want 3", got)
	}
}

func TestS82B10IncubationSacMakesAGolemAsASorcery(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	sac := s82b10Push(g, me.ID, "Incubation Sac", "Artifact", s82b10Sac, 0, 0, 3)
	advanceTo(t, g, game.StepBeginCombat)
	addPool(t, g, me.ID, "{C}{C}{C}{C}")
	if err := s82b10Activate(t, g, me.ID, sac, 0, game.ActivateAbilityParams{}); err == nil {
		t.Fatal("activated outside a main phase")
	}
	toMainForCost(t, g)
	for g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatal(err)
		}
	}
	addPool(t, g, me.ID, "{C}{C}{C}{C}")
	if err := s82b10Activate(t, g, me.ID, sac, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := rfdCountNamed(g, me.ID, "Phyrexian Golem"); got != 1 {
		t.Errorf("%d Golems, want 1", got)
	}
	if got := countersOn(g, sac, oilCounter); got != 2 {
		t.Errorf("%d oil counters left, want 2", got)
	}
}

func TestS82B10FontOfProgressMillsXWhereXIsItsOil(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	font := s82b10Push(g, me.ID, "Font of Progress", "Artifact", s82b10Font, 0, 0, 2)
	lib := opp.Library.Size()
	addPool(t, g, me.ID, "{C}{C}{C}")
	if err := s82b10Activate(t, g, me.ID, font, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := lib - opp.Library.Size(); got != 2 {
		t.Errorf("milled %d, want 2", got)
	}
	if got := countersOn(g, font, oilCounter); got != 2 {
		t.Errorf("the Font spent a counter: %d left, want 2", got)
	}
}

func TestS82B10MeldwebStriderAnimatesForACounterOrACrew(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	v := s82b10Push(g, me.ID, "Meldweb Strider", "Artifact — Vehicle", s82b10Meldweb, 5, 5, 1)
	if c, _ := battlefieldCardByID(g, v); c.IsCreature() {
		t.Fatal("an uncrewed Vehicle is a creature")
	}
	if err := s82b10Activate(t, g, me.ID, v, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("oil animate: %v", err)
	}
	passPriorityAroundTable(t, g)
	c, _ := battlefieldCardByID(g, v)
	if !c.IsCreature() || !c.IsArtifact() {
		t.Error("not an artifact creature after removing the counter")
	}
	if got := countersOn(g, v, oilCounter); got != 0 {
		t.Errorf("%d oil counters left, want 0", got)
	}
	if err := s82b10Activate(t, g, me.ID, v, 0, game.ActivateAbilityParams{}); err == nil {
		t.Error("animated again with no oil counter")
	}
	crewer := pushCrewerForTest(g, me.ID, "Crewer", 3)
	if err := s82b10Activate(t, g, me.ID, v, 1, game.ActivateAbilityParams{CrewIDs: []uuid.UUID{crewer}}); err != nil {
		t.Fatalf("crew 3: %v", err)
	}
}
