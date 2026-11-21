package testsa

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestLifeCardsAreRegistered(t *testing.T) {
	want := map[string]string{
		effectstest.ArchfiendOfDespairOracle: "Archfiend of Despair",
		effectstest.AtarkasCommandOracle:     "Atarka's Command",
		effectstest.ErebosGodOfTheDeadOracle: "Erebos, God of the Dead",
		effectstest.ForsakenWastesOracle:     "Forsaken Wastes",
		effectstest.GiantCindermawOracle:     "Giant Cindermaw",
		effectstest.GrievousWoundOracle:      "Grievous Wound",
		effectstest.GrimaWormtongueOracle:    "Gríma Wormtongue",
		effectstest.HavocFestivalOracle:      "Havoc Festival",
		effectstest.KnightOfDusksOracle:      "Knight of Dusk's Shadow",
		effectstest.QuakebringerOracle:       "Quakebringer",
		effectstest.RampagingFerocidonOracle: "Rampaging Ferocidon",
		effectstest.RoilingVortexOracle:      "Roiling Vortex",
		effectstest.StigmaLasherOracle:       "Stigma Lasher",
		effectstest.TheLordOfPainOracle:      "The Lord of Pain",
		effectstest.TibaltRakishOracle:       "Tibalt, Rakish Instigator",
	}
	for oracle, name := range want {
		spec, ok := effects.Lookup(oracle)
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	cindermaw := effectstest.PushPermanentForTest(g, me.ID, "Giant Cindermaw", effectstest.GiantCindermawOracle, "Creature — Dinosaur Beast")
	if effectstest.Gains(t, g, me.ID) || effectstest.Gains(t, g, opp.ID) {
		t.Fatal("Giant Cindermaw: nobody gains life")
	}
	if _, err := g.Battlefield.Remove(cindermaw); err != nil {
		t.Fatal(err)
	}
	effectstest.PushPermanentForTest(g, me.ID, "Knight of Dusk's Shadow", effectstest.KnightOfDusksOracle, "Creature — Human Knight")
	if !effectstest.Gains(t, g, me.ID) {
		t.Error("Knight of Dusk's Shadow: its controller still gains life")
	}
	if effectstest.Gains(t, g, opp.ID) {
		t.Error("Knight of Dusk's Shadow: an opponent gains no life")
	}
}

// Erebos isn't a creature without five devotion, and its static works
// either way.
func TestErebosGodOfTheDead(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	erebos := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Erebos, God of the Dead", OracleID: effectstest.ErebosGodOfTheDeadOracle,
		TypeLine: "Legendary Enchantment Creature — God", ManaCost: "{3}{B}", Power: 5, Toughness: 7,
		Owner: me.ID, Controller: me.ID,
	})
	if effectstest.B39IsCreature(t, g, erebos) {
		t.Error("devotion 1: Erebos isn't a creature")
	}
	if effectstest.Gains(t, g, opp.ID) || !effectstest.Gains(t, g, me.ID) {
		t.Error("Erebos: opponents can't gain life, its controller can")
	}
}

// Grievous Wound: the enchanted player can't gain life and loses half
// their life, rounded up, when dealt damage.
func TestGrievousWound(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	wound := effectstest.PushPermanentForTest(g, me.ID, "Grievous Wound", effectstest.GrievousWoundOracle, "Enchantment — Aura")
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == wound {
			g.Battlefield.Cards[i].AttachedTo = game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}
		}
	}
	if effectstest.Gains(t, g, opp.ID) || !effectstest.Gains(t, g, me.ID) {
		t.Fatal("only the enchanted player can't gain life")
	}
	start := opp.Life // 40
	g.WithWriteLock(func() {
		if err := g.DealDamageToPlayerForEffect(uuid.Nil, opp.ID, 1); err != nil {
			t.Fatal(err)
		}
	})
	effectstest.PassPriorityAroundTable(t, g)
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
		{"Havoc Festival", effectstest.HavocFestivalOracle, "Enchantment", func(l int) int { return l - (l+1)/2 }},
		{"Forsaken Wastes", effectstest.ForsakenWastesOracle, "World Enchantment", func(l int) int { return l - 1 }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			effectstest.PushPermanentForTest(g, me.ID, tc.name, tc.oracle, tc.typeLine)
			if effectstest.Gains(t, g, me.ID) || effectstest.Gains(t, g, opp.ID) {
				t.Fatal("players can't gain life")
			}
			start := opp.Life
			effectstest.AdvanceToUpkeepOf(t, g, 1)
			effectstest.PassPriorityAroundTable(t, g)
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushPermanentForTest(g, me.ID, "Sulfuric Vortex", effectstest.SulfuricVortexOracle, "Enchantment")
	if effectstest.Gains(t, g, me.ID) || effectstest.Gains(t, g, opp.ID) {
		t.Fatal("Sulfuric Vortex: no player gains life")
	}
	start := opp.Life
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != start-2 {
		t.Errorf("opponent's upkeep: life %d, want %d", opp.Life, start-2)
	}
	spec, _ := effects.Lookup(effectstest.SulfuricVortexOracle)
	if spec.Completeness != effects.CompletenessFull {
		t.Error("Sulfuric Vortex is complete")
	}
}

// Archfiend of Despair: each opponent loses the life they lost this
// turn at each end step.
func TestArchfiendOfDespairEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushPermanentForTest(g, me.ID, "Archfiend of Despair", effectstest.ArchfiendOfDespairOracle, "Creature — Demon")
	if effectstest.Gains(t, g, opp.ID) {
		t.Fatal("an opponent can't gain life")
	}
	if _, err := g.ChangePlayerLife(opp.ID, -5); err != nil {
		t.Fatal(err)
	}
	start := opp.Life
	effectstest.AdvanceToEndStepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.CastCatalogSpellWithModes(t, g, "Atarka's Command", "Instant", effectstest.AtarkasCommandOracle, []int{0, 1})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != 37 {
		t.Errorf("opponent at %d, want 37", opp.Life)
	}
	if effectstest.Gains(t, g, opp.ID) || !effectstest.Gains(t, g, me.ID) {
		t.Error("this turn, opponents can't gain life and the caster can")
	}
}

// Screaming Nemesis: a player it damages can't gain life for the rest
// of the game.
func TestScreamingNemesisStopsLifeGainForTheRestOfTheGame(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	nemesis := effectstest.PushPermanentForTest(g, me.ID, "Screaming Nemesis", effectstest.ScreamingNemesisOracle, "Creature — Spirit")
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventDealDamage, Source: uuid.Nil, Target: nemesis, Amount: 2})
	})
	g.RunStateChecksForTest() // #3172: "dealt damage" triggers wait for the batch to settle
	effectstest.AnswerPickTargetPlayer(t, g, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != 38 {
		t.Fatalf("opponent at %d, want 38", opp.Life)
	}
	if effectstest.Gains(t, g, opp.ID) {
		t.Error("the damaged player can't gain life")
	}
	if !effectstest.Gains(t, g, me.ID) {
		t.Error("only the damaged player is covered")
	}
	spec, _ := effects.Lookup(effectstest.ScreamingNemesisOracle)
	if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("Nemesis is full now, got %v %v", spec.Completeness, spec.Caveats)
	}
}

// Stigma Lasher: a player it damages can't gain life for the rest of
// the game.
func TestStigmaLasherStopsLifeGain(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	lasher := effectstest.PushPermanentForTest(g, me.ID, "Stigma Lasher", effectstest.StigmaLasherOracle, "Creature — Elemental Shaman")
	g.WithWriteLock(func() {
		if err := g.DealDamageToPlayerForEffect(lasher, opp.ID, 2); err != nil {
			t.Fatal(err)
		}
	})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.Gains(t, g, opp.ID) {
		t.Error("the damaged player can't gain life")
	}
	if !effectstest.Gains(t, g, g.Seats[2].ID) {
		t.Error("an undamaged player still gains life")
	}
}

// Rampaging Ferocidon pings the controller of another creature that
// enters.
func TestRampagingFerocidonPingsTheController(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushPermanentForTest(g, g.Seats[1].ID, "Rampaging Ferocidon", effectstest.RampagingFerocidonOracle, "Creature — Dinosaur")
	effectstest.CastCatalogSpell(t, g, "Grizzly Bears", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != 39 {
		t.Errorf("the entering creature's controller is at %d, want 39", me.Life)
	}
}

// The Lord of Pain: the first spell a player casts each turn deals its
// mana value to another target player.
func TestTheLordOfPainFirstSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushPermanentForTest(g, opp.ID, "The Lord of Pain", effectstest.TheLordOfPainOracle, "Legendary Creature — Human Assassin")
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{InstanceID: id, Name: "Three Drop", TypeLine: "Sorcery", ManaCost: "{2}{R}",
		Owner: me.ID, Controller: me.ID})
	effectstest.AdvanceToMain(t, g)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatal(err)
	}
	effectstest.AnswerPickTargetPlayer(t, g, g.Seats[2].ID)
	effectstest.PassPriorityAroundTable(t, g)
	if got := g.Seats[2].Life; got != 37 {
		t.Errorf("the chosen player is at %d, want 37", got)
	}
}

// Quakebringer's upkeep trigger works from the graveyard only while its
// owner controls a Giant.
func TestQuakebringerFromTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	owner := g.Seats[0]
	owner.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Quakebringer", OracleID: effectstest.QuakebringerOracle,
		TypeLine: "Creature — Giant Berserker", Owner: owner.ID, Controller: owner.ID})
	start := opp.Life
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != start {
		t.Fatalf("no Giant: the graveyard Quakebringer dealt damage (%d -> %d)", start, opp.Life)
	}
	effectstest.PushPermanentForTest(g, owner.ID, "Hill Giant", "", "Creature — Giant")
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != start-2 {
		t.Errorf("with a Giant: opponent at %d, want %d", opp.Life, start-2)
	}
}

// Roiling Vortex's activation: opponents can't gain life this turn.
func TestRoilingVortexActivation(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	vortex := effectstest.PushPermanentForTest(g, me.ID, "Roiling Vortex", effectstest.RoilingVortexOracle, "Enchantment")
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "R")
	effectstest.B16Activate(t, g, me.ID, vortex, 0, game.ActivateAbilityParams{})
	if effectstest.Gains(t, g, opp.ID) || !effectstest.Gains(t, g, me.ID) {
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
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			grima := effectstest.PushCatalogPermanent(g, me.ID, "Gríma Wormtongue", "Legendary Creature — Human Advisor", effectstest.GrimaWormtongueOracle, false)
			fodder := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
				InstanceID: uuid.New(), Name: "Fodder", TypeLine: tc.supertype + "Creature — Human",
				Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
			})
			effectstest.AdvanceToMain(t, g)
			effectstest.B16Activate(t, g, me.ID, grima, 0, game.ActivateAbilityParams{
				SacrificeIDs: []uuid.UUID{fodder},
				Targets:      []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
			})
			if opp.Life != 39 {
				t.Errorf("target at %d, want 39", opp.Life)
			}
			if got := effectstest.B16CountNamed(g, "Orc Army") > 0; got != tc.army {
				t.Errorf("Orc Army amassed = %v, want %v", got, tc.army)
			}
		})
	}
}

// Tibalt's −2 makes a Devil; his static stops opponents gaining life.
func TestTibaltRakishInstigator(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	tibalt := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Tibalt, Rakish Instigator", OracleID: effectstest.TibaltRakishOracle,
		TypeLine: "Legendary Planeswalker — Tibalt", Counters: map[string]int{game.CounterLoyalty: 5},
		Owner: me.ID, Controller: me.ID,
	})
	if effectstest.Gains(t, g, opp.ID) {
		t.Error("an opponent can't gain life")
	}
	effectstest.AdvanceToMain(t, g)
	effectstest.B16Activate(t, g, me.ID, tibalt, 0, game.ActivateAbilityParams{})
	if effectstest.B16CountNamed(g, "Devil") != 1 {
		t.Error("the −2 makes a Devil")
	}
}
