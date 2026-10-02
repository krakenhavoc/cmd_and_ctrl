package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// adr0107_pr6_life_cards_test.go — ADR 0107 §5 (#1880): the "can't
// gain life" cards, and Sulfuric Vortex's and Screaming Nemesis's
// clauses that waited on the same seam.

const (
	archfiendOfDespairOracle = "9c36760b-57c5-488b-afc9-ef141942c6ab"
	atarkasCommandOracle     = "1943cbb0-2b12-4450-90af-060b8c8627c3"
	erebosGodOfTheDeadOracle = "1cfc7b12-a595-492d-81ee-bd100ba7de6a"
	forsakenWastesOracle     = "b9e61e68-9dc8-4295-95dc-dd66a0907c8c"
	giantCindermawOracle     = "bd0c0e85-ff24-46c9-a547-4f633355c131"
	grievousWoundOracle      = "4d97b9d1-1138-40da-b593-8065ea6ed7b3"
	grimaWormtongueOracle    = "8f98806e-c6b7-44af-b436-88acf36d25ec"
	havocFestivalOracle      = "af0dd100-67f0-4d7c-a65b-b2c7149288f7"
	knightOfDusksOracle      = "815ae30f-f455-4795-96eb-4bcc41d415e8"
	quakebringerOracle       = "bbe41a21-f8ad-4e5c-88d8-2a41dfd32d13"
	rampagingFerocidonOracle = "3e5ca524-8dd1-4f7f-a467-5210d768b8a1"
	roilingVortexOracle      = "34f83b7a-45a7-40a8-abcb-31d7aae4cea1"
	stigmaLasherOracle       = "7e4a1286-cac6-491f-a32d-64fb98d35de2"
	theLordOfPainOracle      = "b194f277-5045-4391-8b48-d555a6d657a7"
	tibaltRakishOracle       = "0b768f8f-2213-45b9-bced-8fb1bbb441c3"
	sulfuricVortexOracle     = "7652f328-e142-494b-a869-772ced10c26a"
	screamingNemesisOracle   = "fcb7c93c-46ab-49b5-a6e0-35d73f3be8f0"
)

// gains reports whether `player` gains life from a 3-point sandbox gain,
// leaving the total where it was found.
func gains(t *testing.T, g *game.Game, player uuid.UUID) bool {
	t.Helper()
	before := lifeTotalOf(g, player)
	if _, err := g.ChangePlayerLife(player, 3); err != nil {
		t.Fatalf("ChangePlayerLife: %v", err)
	}
	after := lifeTotalOf(g, player)
	if after != before {
		if _, err := g.ChangePlayerLife(player, before-after); err != nil {
			t.Fatalf("ChangePlayerLife: %v", err)
		}
	}
	return after > before
}

func TestLifeCardsAreRegistered(t *testing.T) {
	want := map[string]string{
		archfiendOfDespairOracle: "Archfiend of Despair",
		atarkasCommandOracle:     "Atarka's Command",
		erebosGodOfTheDeadOracle: "Erebos, God of the Dead",
		forsakenWastesOracle:     "Forsaken Wastes",
		giantCindermawOracle:     "Giant Cindermaw",
		grievousWoundOracle:      "Grievous Wound",
		grimaWormtongueOracle:    "Gríma Wormtongue",
		havocFestivalOracle:      "Havoc Festival",
		knightOfDusksOracle:      "Knight of Dusk's Shadow",
		quakebringerOracle:       "Quakebringer",
		rampagingFerocidonOracle: "Rampaging Ferocidon",
		roilingVortexOracle:      "Roiling Vortex",
		stigmaLasherOracle:       "Stigma Lasher",
		theLordOfPainOracle:      "The Lord of Pain",
		tibaltRakishOracle:       "Tibalt, Rakish Instigator",
	}
	for oracle, name := range want {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name {
			t.Errorf("oracle %s registered as %q, want %q", oracle, spec.Name, name)
		}
	}
}

// "Players can't gain life" covers everyone, the controller included;
// "your opponents can't gain life" spares the controller.
func TestPlayersAndOpponentsCantGainLifeStatics(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	cindermaw := pushPermanentForTest(g, me.ID, "Giant Cindermaw", giantCindermawOracle, "Creature — Dinosaur Beast")
	if gains(t, g, me.ID) || gains(t, g, opp.ID) {
		t.Fatal("Giant Cindermaw: nobody gains life")
	}
	if _, err := g.Battlefield.Remove(cindermaw); err != nil {
		t.Fatal(err)
	}
	pushPermanentForTest(g, me.ID, "Knight of Dusk's Shadow", knightOfDusksOracle, "Creature — Human Knight")
	if !gains(t, g, me.ID) {
		t.Error("Knight of Dusk's Shadow: its controller still gains life")
	}
	if gains(t, g, opp.ID) {
		t.Error("Knight of Dusk's Shadow: an opponent gains no life")
	}
}

// Erebos isn't a creature without five devotion, and its static works
// either way.
func TestErebosGodOfTheDead(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	erebos := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Erebos, God of the Dead", OracleID: erebosGodOfTheDeadOracle,
		TypeLine: "Legendary Enchantment Creature — God", ManaCost: "{3}{B}", Power: 5, Toughness: 7,
		Owner: me.ID, Controller: me.ID,
	})
	if b39IsCreature(t, g, erebos) {
		t.Error("devotion 1: Erebos isn't a creature")
	}
	if gains(t, g, opp.ID) || !gains(t, g, me.ID) {
		t.Error("Erebos: opponents can't gain life, its controller can")
	}
}

// Grievous Wound: the enchanted player can't gain life and loses half
// their life, rounded up, when dealt damage.
func TestGrievousWound(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	wound := pushPermanentForTest(g, me.ID, "Grievous Wound", grievousWoundOracle, "Enchantment — Aura")
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == wound {
			g.Battlefield.Cards[i].AttachedTo = game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}
		}
	}
	if gains(t, g, opp.ID) || !gains(t, g, me.ID) {
		t.Fatal("only the enchanted player can't gain life")
	}
	start := opp.Life // 40
	g.WithWriteLock(func() {
		if err := g.DealDamageToPlayerForEffect(uuid.Nil, opp.ID, 1); err != nil {
			t.Fatal(err)
		}
	})
	passPriorityAroundTable(t, g)
	after := start - 1
	if want := after - (after+1)/2; opp.Life != want {
		t.Errorf("enchanted player at %d after 1 damage and the halving, want %d", opp.Life, want)
	}
}

// Havoc Festival and Forsaken Wastes hit whoever's upkeep it is.
func TestUpkeepLifeLossForEachPlayer(t *testing.T) {
	for _, tc := range []struct {
		name, oracle, typeLine string
		want                   func(life int) int
	}{
		{"Havoc Festival", havocFestivalOracle, "Enchantment", func(l int) int { return l - (l+1)/2 }},
		{"Forsaken Wastes", forsakenWastesOracle, "World Enchantment", func(l int) int { return l - 1 }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			pushPermanentForTest(g, me.ID, tc.name, tc.oracle, tc.typeLine)
			if gains(t, g, me.ID) || gains(t, g, opp.ID) {
				t.Fatal("players can't gain life")
			}
			start := opp.Life
			advanceToUpkeepOf(t, g, 1)
			passPriorityAroundTable(t, g)
			if want := tc.want(start); opp.Life != want {
				t.Errorf("opponent at %d after their upkeep, want %d", opp.Life, want)
			}
			if me.Life != 40 {
				t.Errorf("the controller's life moved on an opponent's upkeep: %d", me.Life)
			}
		})
	}
}

// Sulfuric Vortex: no player gains life, and it hits every player's
// upkeep, not only its controller's.
func TestSulfuricVortexGainsNoLifeAndHitsEachUpkeep(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushPermanentForTest(g, me.ID, "Sulfuric Vortex", sulfuricVortexOracle, "Enchantment")
	if gains(t, g, me.ID) || gains(t, g, opp.ID) {
		t.Fatal("Sulfuric Vortex: no player gains life")
	}
	start := opp.Life
	advanceToUpkeepOf(t, g, 1)
	passPriorityAroundTable(t, g)
	if opp.Life != start-2 {
		t.Errorf("opponent's upkeep: life %d, want %d", opp.Life, start-2)
	}
	spec, _ := Lookup(sulfuricVortexOracle)
	if spec.Completeness != CompletenessFull {
		t.Error("Sulfuric Vortex is complete")
	}
}

// Archfiend of Despair: each opponent loses the life they lost this
// turn at each end step.
func TestArchfiendOfDespairEndStep(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushPermanentForTest(g, me.ID, "Archfiend of Despair", archfiendOfDespairOracle, "Creature — Demon")
	if gains(t, g, opp.ID) {
		t.Fatal("an opponent can't gain life")
	}
	if _, err := g.ChangePlayerLife(opp.ID, -5); err != nil {
		t.Fatal(err)
	}
	start := opp.Life
	advanceToEndStepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	if opp.Life != start-5 {
		t.Errorf("end step: opponent at %d, want %d", opp.Life, start-5)
	}
	if me.Life != 40 {
		t.Errorf("the controller lost life: %d", me.Life)
	}
}

// Atarka's Command: opponents can't gain life this turn and take 3; its
// caster can still gain.
func TestAtarkasCommandLifeAndDamageModes(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	castCatalogSpellWithModes(t, g, "Atarka's Command", "Instant", atarkasCommandOracle, []int{0, 1})
	passPriorityAroundTable(t, g)
	if opp.Life != 37 {
		t.Errorf("opponent at %d, want 37", opp.Life)
	}
	if gains(t, g, opp.ID) || !gains(t, g, me.ID) {
		t.Error("this turn, opponents can't gain life and the caster can")
	}
}

// Screaming Nemesis: a player it damages can't gain life for the rest
// of the game.
func TestScreamingNemesisStopsLifeGainForTheRestOfTheGame(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	nemesis := pushPermanentForTest(g, me.ID, "Screaming Nemesis", screamingNemesisOracle, "Creature — Spirit")
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventDealDamage, Source: uuid.Nil, Target: nemesis, Amount: 2})
	})
	answerPickTargetPlayer(t, g, opp.ID)
	passPriorityAroundTable(t, g)
	if opp.Life != 38 {
		t.Fatalf("opponent at %d, want 38", opp.Life)
	}
	if gains(t, g, opp.ID) {
		t.Error("the damaged player can't gain life")
	}
	if !gains(t, g, me.ID) {
		t.Error("only the damaged player is covered")
	}
	spec, _ := Lookup(screamingNemesisOracle)
	if len(spec.Caveats) != 1 {
		t.Errorf("one caveat left (simultaneous damage), got %v", spec.Caveats)
	}
}

// Stigma Lasher: a player it damages can't gain life for the rest of
// the game.
func TestStigmaLasherStopsLifeGain(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	lasher := pushPermanentForTest(g, me.ID, "Stigma Lasher", stigmaLasherOracle, "Creature — Elemental Shaman")
	g.WithWriteLock(func() {
		if err := g.DealDamageToPlayerForEffect(lasher, opp.ID, 2); err != nil {
			t.Fatal(err)
		}
	})
	passPriorityAroundTable(t, g)
	if gains(t, g, opp.ID) {
		t.Error("the damaged player can't gain life")
	}
	if !gains(t, g, g.Seats[2].ID) {
		t.Error("an undamaged player still gains life")
	}
}

// Rampaging Ferocidon pings the controller of another creature that
// enters.
func TestRampagingFerocidonPingsTheController(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushPermanentForTest(g, g.Seats[1].ID, "Rampaging Ferocidon", rampagingFerocidonOracle, "Creature — Dinosaur")
	castCatalogSpell(t, g, "Grizzly Bears", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	if me.Life != 39 {
		t.Errorf("the entering creature's controller is at %d, want 39", me.Life)
	}
}

// The Lord of Pain: the first spell a player casts each turn deals its
// mana value to another target player.
func TestTheLordOfPainFirstSpell(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushPermanentForTest(g, opp.ID, "The Lord of Pain", theLordOfPainOracle, "Legendary Creature — Human Assassin")
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{InstanceID: id, Name: "Three Drop", TypeLine: "Sorcery", ManaCost: "{2}{R}",
		Owner: me.ID, Controller: me.ID})
	advanceToMain(t, g)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatal(err)
	}
	answerPickTargetPlayer(t, g, g.Seats[2].ID)
	passPriorityAroundTable(t, g)
	if got := g.Seats[2].Life; got != 37 {
		t.Errorf("the chosen player is at %d, want 37", got)
	}
}

// Quakebringer's upkeep trigger works from the graveyard only while its
// owner controls a Giant.
func TestQuakebringerFromTheGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	owner := g.Seats[0]
	owner.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Quakebringer", OracleID: quakebringerOracle,
		TypeLine: "Creature — Giant Berserker", Owner: owner.ID, Controller: owner.ID})
	start := opp.Life
	advanceToUpkeepOf(t, g, 1)
	advanceToUpkeepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	if opp.Life != start {
		t.Fatalf("no Giant: the graveyard Quakebringer dealt damage (%d -> %d)", start, opp.Life)
	}
	pushPermanentForTest(g, owner.ID, "Hill Giant", "", "Creature — Giant")
	advanceToUpkeepOf(t, g, 1)
	advanceToUpkeepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	if opp.Life != start-2 {
		t.Errorf("with a Giant: opponent at %d, want %d", opp.Life, start-2)
	}
}

// Roiling Vortex's activation: opponents can't gain life this turn.
func TestRoilingVortexActivation(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	vortex := pushPermanentForTest(g, me.ID, "Roiling Vortex", roilingVortexOracle, "Enchantment")
	advanceToMain(t, g)
	b06AddMana(me, "R")
	b16Activate(t, g, me.ID, vortex, 0, game.ActivateAbilityParams{})
	if gains(t, g, opp.ID) || !gains(t, g, me.ID) {
		t.Error("opponents can't gain life this turn; the activator can")
	}
}

// Gríma Wormtongue: the target loses 1, and a legendary sacrifice amasses
// Orcs 2.
func TestGrimaWormtongue(t *testing.T) {
	for _, tc := range []struct {
		name      string
		supertype string
		army      bool
	}{{"legendary", "Legendary ", true}, {"ordinary", "", false}} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			grima := pushCatalogPermanent(g, me.ID, "Gríma Wormtongue", "Legendary Creature — Human Advisor", grimaWormtongueOracle, false)
			fodder := pushBattlefieldCardWithTimestamp(g, game.Card{
				InstanceID: uuid.New(), Name: "Fodder", TypeLine: tc.supertype + "Creature — Human",
				Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
			})
			advanceToMain(t, g)
			b16Activate(t, g, me.ID, grima, 0, game.ActivateAbilityParams{
				SacrificeIDs: []uuid.UUID{fodder},
				Targets:      []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
			})
			if opp.Life != 39 {
				t.Errorf("target at %d, want 39", opp.Life)
			}
			if got := b16CountNamed(g, "Orc Army") > 0; got != tc.army {
				t.Errorf("Orc Army amassed = %v, want %v", got, tc.army)
			}
		})
	}
}

// Tibalt's −2 makes a Devil; his static stops opponents gaining life.
func TestTibaltRakishInstigator(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	tibalt := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Tibalt, Rakish Instigator", OracleID: tibaltRakishOracle,
		TypeLine: "Legendary Planeswalker — Tibalt", Counters: map[string]int{game.CounterLoyalty: 5},
		Owner: me.ID, Controller: me.ID,
	})
	if gains(t, g, opp.ID) {
		t.Error("an opponent can't gain life")
	}
	advanceToMain(t, g)
	b16Activate(t, g, me.ID, tibalt, 0, game.ActivateAbilityParams{})
	if b16CountNamed(g, "Devil") != 1 {
		t.Error("the −2 makes a Devil")
	}
}
