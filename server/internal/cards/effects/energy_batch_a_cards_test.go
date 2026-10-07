package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// energy_batch_a_cards_test.go — ADR 0129 PR 1, batch A: the cards that
// get energy and spend none, plus Maximus's sacrifice outlet.

const (
	eaAttune         = "b337c241-f2c4-48e3-a303-371b172fcd18"
	eaMeltdown       = "433a2880-d989-4c40-99fa-0896de341b88"
	eaElectrosiphon  = "bff9105f-d8d4-4412-88d9-c9697581ebe9"
	eaVoyager        = "e7112d60-1154-4770-9ec4-c7038157be8a"
	eaGlassblower    = "0fbe0610-8565-4910-bf35-2c10e38b633a"
	eaGlimmer        = "b9c58cfb-b7d5-4e86-a5d3-efb07192aa22"
	eaHighspire      = "c86ed70b-ab09-402e-8aaf-6332b4749453"
	eaInventor       = "126679aa-f150-4949-8566-12d223af3960"
	eaLiveFast       = "04d02394-e17f-43cd-932f-f58e8cc56629"
	eaReservoir      = "7177361a-f86f-4a50-b517-229cb1eaac07"
	eaRogueRefiner   = "b2f09e41-0a91-4fb6-8804-874b3a5166b0"
	eaSage           = "a0b413cc-b16f-4112-8061-2ccded7f481d"
	eaTune           = "4262a726-d1da-4bdc-ab14-73212aa83f9c"
	eaWoodweaver     = "3334b218-01e3-47a3-8a28-5cc2decbedea"
	eaMaximus        = "23f0d43c-f84d-48cf-8ee7-bfda6e0f7e29"
	eaBasicLandMatch = "Forest"
)

// Every batch A card is registered Full.
func TestEnergyBatchACardsAreFull(t *testing.T) {
	for _, oracle := range []string{eaAttune, eaMeltdown, eaElectrosiphon, eaVoyager, eaGlassblower, eaGlimmer,
		eaHighspire, eaInventor, eaLiveFast, eaReservoir, eaRogueRefiner, eaSage, eaTune, eaWoodweaver, eaMaximus} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", oracle)
			continue
		}
		if spec.Completeness != CompletenessFull {
			t.Errorf("%s: completeness %v, want Full", spec.Name, spec.Completeness)
		}
	}
}

// The enters-the-battlefield energy cards: each gives its printed energy,
// and Reservoir Walker and Rogue Refiner do the rest of their line too.
func TestEnergyBatchAEntersGiveEnergy(t *testing.T) {
	for _, tc := range []struct {
		name, oracle, typeLine string
		energy                 int
	}{
		{"Sage of Shaila's Claim", eaSage, "Creature — Elf Druid", 3},
		{"Reservoir Walker", eaReservoir, "Artifact Creature — Construct", 3},
		{"Rogue Refiner", eaRogueRefiner, "Creature — Human Rogue", 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[0]
			life, hand := me.Life, me.Hand.Size()
			castAndResolveCreature(t, g, tc.name, tc.typeLine, tc.oracle)
			passPriorityAroundTable(t, g)
			if got := energyOf(me); got != tc.energy {
				t.Errorf("energy = %d, want %d", got, tc.energy)
			}
			switch tc.oracle {
			case eaReservoir:
				if me.Life != life+3 {
					t.Errorf("life = %d, want %d", me.Life, life+3)
				}
			case eaRogueRefiner:
				if me.Hand.Size() != hand+1 {
					t.Errorf("hand = %d, want %d (one drawn)", me.Hand.Size(), hand+1)
				}
			}
		})
	}
}

func TestEnergyBatchAAttuneWithAether(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedSearchLibrary(me, searchTestLand(eaBasicLandMatch, "Basic Land — Forest"), searchTestLand("Odd", "Sorcery"))
	castCatalogSpell(t, g, "Attune with Aether", "Sorcery", eaAttune, nil)
	passPriorityAroundTable(t, g)
	if searchChoiceFor(g, me.ID) != nil {
		answerSearchNamed(t, g, me.ID, eaBasicLandMatch)
		passPriorityAroundTable(t, g)
	}
	found := false
	for _, c := range me.Hand.Cards {
		if c.Name == eaBasicLandMatch {
			found = true
		}
	}
	if !found {
		t.Error("the Forest is not in hand")
	}
	if energyOf(me) != 2 {
		t.Errorf("energy = %d, want 2", energyOf(me))
	}
}

func TestEnergyBatchAAetherMeltdownShrinksAndGivesEnergy(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := pushVanillaCreature(g, opp.ID, "Victim", 5, 5)
	castCatalogSpell(t, g, "Aether Meltdown", "Enchantment — Aura", eaMeltdown,
		[]game.TargetRef{{Kind: game.TargetCard, ID: victim}})
	passPriorityAroundTable(t, g)
	if energyOf(me) != 2 {
		t.Errorf("energy = %d, want 2", energyOf(me))
	}
	if got := apaLive(g, victim).CurrentPower(); got != 1 {
		t.Errorf("enchanted creature's power = %d, want 1", got)
	}
}

func TestEnergyBatchAElectrosiphonCountersAndGivesItsManaValue(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	spell := uuid.New()
	opp.Hand.PushTop(game.Card{InstanceID: spell, Name: "Big Instant", TypeLine: "Instant", ManaCost: "{3}{R}",
		Owner: opp.ID, Controller: opp.ID})
	if err := g.CastSpell(opp.ID, spell, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast the target: %v", err)
	}
	siphon := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: siphon, Name: "Electrosiphon", TypeLine: "Instant", OracleID: eaElectrosiphon,
		Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, siphon, game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: spell}}}); err != nil {
		t.Fatalf("cast Electrosiphon: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(spell) {
		t.Error("the target was not countered")
	}
	if energyOf(me) != 4 {
		t.Errorf("energy = %d, want 4 (its mana value)", energyOf(me))
	}
}

func TestEnergyBatchAEmpyrealVoyagerGetsThatMany(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	v := pushDiesCreatureForTest(g, me.ID, "Empyreal Voyager", eaVoyager, "Creature — Vedalken Scout", 2, 3)
	attackWith(t, g, opp.ID, v)
	passPriorityAroundTable(t, g)
	if energyOf(me) != 2 {
		t.Errorf("energy = %d, want 2 (the damage dealt)", energyOf(me))
	}
}

func TestEnergyBatchAGlassblowersPuzzleknot(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedLibrary(me, "A", "B", "C")
	knot := castAndResolveCreature(t, g, "Glassblower's Puzzleknot", "Artifact", eaGlassblower)
	passPriorityAroundTable(t, g)
	if energyOf(me) != 0 {
		t.Fatalf("energy before the scry is answered = %d, want 0", energyOf(me))
	}
	answerScryKeepAll(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if energyOf(me) != 2 {
		t.Fatalf("energy after the enters scry = %d, want 2", energyOf(me))
	}
	if err := g.ActivateCatalogAbility(me.ID, knot, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(knot) {
		t.Error("the Puzzleknot was not sacrificed")
	}
	passPriorityAroundTable(t, g)
	answerScryKeepAll(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if energyOf(me) != 4 {
		t.Errorf("energy after the activation = %d, want 4", energyOf(me))
	}
}

func TestEnergyBatchAWoodweaversPuzzleknot(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	life := me.Life
	knot := castAndResolveCreature(t, g, "Woodweaver's Puzzleknot", "Artifact", eaWoodweaver)
	passPriorityAroundTable(t, g)
	if energyOf(me) != 3 || me.Life != life+3 {
		t.Fatalf("after entering: energy %d life %d, want 3 and %d", energyOf(me), me.Life, life+3)
	}
	if err := g.ActivateCatalogAbility(me.ID, knot, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if energyOf(me) != 6 || me.Life != life+6 {
		t.Errorf("after the activation: energy %d life %d, want 6 and %d", energyOf(me), me.Life, life+6)
	}
}

func TestEnergyBatchAGlimmerOfGenius(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedLibrary(me, "A", "B", "C", "D")
	hand := me.Hand.Size()
	castCatalogSpell(t, g, "Glimmer of Genius", "Instant", eaGlimmer, nil)
	passPriorityAroundTable(t, g)
	answerScryKeepAll(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+2 {
		t.Errorf("hand = %d, want %d", me.Hand.Size(), hand+2)
	}
	if energyOf(me) != 2 {
		t.Errorf("energy = %d, want 2", energyOf(me))
	}
}

func TestEnergyBatchAHighspireInfusion(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	castCatalogSpell(t, g, "Highspire Infusion", "Instant", eaHighspire, []game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	passPriorityAroundTable(t, g)
	if got := apaLive(g, bear).CurrentPower(); got != 5 {
		t.Errorf("power = %d, want 5", got)
	}
	if energyOf(me) != 2 {
		t.Errorf("energy = %d, want 2", energyOf(me))
	}
}

func TestEnergyBatchAInspiredInventorModes(t *testing.T) {
	t.Run("energy", func(t *testing.T) {
		g := newCatalogGame(t)
		me := g.Seats[0]
		castAndResolveCreature(t, g, "Inspired Inventor", "Creature — Human Artificer", eaInventor)
		c := latestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID)
		if c == nil {
			t.Fatal("no mode_pick prompt")
		}
		if err := g.ResolveModePick(c.ID, me.ID, []int{0}); err != nil {
			t.Fatal(err)
		}
		passPriorityAroundTable(t, g)
		if energyOf(me) != 3 {
			t.Errorf("energy = %d, want 3", energyOf(me))
		}
	})
	t.Run("counter", func(t *testing.T) {
		g := newCatalogGame(t)
		me := g.Seats[0]
		bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
		castAndResolveCreature(t, g, "Inspired Inventor", "Creature — Human Artificer", eaInventor)
		c := latestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID)
		if c == nil {
			t.Fatal("no mode_pick prompt")
		}
		if err := g.ResolveModePick(c.ID, me.ID, []int{1}); err != nil {
			t.Fatal(err)
		}
		pickCard(t, g, me.ID, bear)
		passPriorityAroundTable(t, g)
		if n := apaLive(g, bear).Counters[game.CounterPlusOne]; n != 1 {
			t.Errorf("+1/+1 counters = %d, want 1", n)
		}
		if energyOf(me) != 0 {
			t.Errorf("energy = %d, want 0", energyOf(me))
		}
	})
}

func TestEnergyBatchALiveFastAndTuneTheNarrative(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	life, hand := me.Life, me.Hand.Size()
	castCatalogSpell(t, g, "Live Fast", "Sorcery", eaLiveFast, nil)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+2 || me.Life != life-2 || energyOf(me) != 2 {
		t.Errorf("Live Fast: hand %d life %d energy %d, want %d, %d and 2", me.Hand.Size(), me.Life, energyOf(me), hand+2, life-2)
	}
	hand = me.Hand.Size()
	castCatalogSpell(t, g, "Tune the Narrative", "Instant", eaTune, nil)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 || energyOf(me) != 4 {
		t.Errorf("Tune the Narrative: hand %d energy %d, want %d and 4", me.Hand.Size(), energyOf(me), hand+1)
	}
}

func TestEnergyBatchAMaximusSacrificesAnArtifactForEnergy(t *testing.T) {
	g, me, _ := p7Table(t)
	maximus := apaPush(g, me.ID, me.ID, game.Card{Name: "Maximus, Knight Apparent", OracleID: eaMaximus,
		TypeLine: "Legendary Creature — Human Knight", Power: 4, Toughness: 4})
	if err := g.ActivateCatalogAbility(me.ID, maximus, 0, game.ActivateAbilityParams{}); err == nil {
		t.Fatal("activated with no artifact to sacrifice")
	}
	rock := apaPush(g, me.ID, me.ID, game.Card{Name: "Rock", TypeLine: "Artifact"})
	p7Activate(t, g, me, maximus, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{rock}})
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact was not sacrificed")
	}
	if energyOf(me) != 2 {
		t.Errorf("energy = %d, want 2", energyOf(me))
	}
	if e := specActivatedEnergy(t, eaMaximus, 0); e.Energy != 0 || e.Mana != "{1}" {
		t.Errorf("Maximus's cost = %+v, want {1} and a sacrifice", e)
	}
}

// Maximus's enters search is optional and finds only an Equipment card
// with mana value 2.
func TestEnergyBatchAMaximusSearchesForATwoDropEquipment(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedSearchLibrary(me,
		game.Card{Name: "Two Drop", TypeLine: "Artifact — Equipment", ManaCost: "{2}"},
		game.Card{Name: "Three Drop", TypeLine: "Artifact — Equipment", ManaCost: "{3}"})
	castAndResolveCreature(t, g, "Maximus, Knight Apparent", "Legendary Creature — Human Knight", eaMaximus)
	passPriorityAroundTable(t, g)
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	if searchOptionNamed(g, c, "Three Drop") != uuid.Nil {
		t.Error("the search offered a mana value 3 Equipment")
	}
	answerSearchNamed(t, g, me.ID, "Two Drop")
}
