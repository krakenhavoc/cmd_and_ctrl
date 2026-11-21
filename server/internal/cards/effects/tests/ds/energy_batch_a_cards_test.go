package testsds

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Every batch A card is registered Full.
func TestEnergyBatchACardsAreFull(t *testing.T) {
	for _, oracle := range []string{effectstest.EaAttune, effectstest.EaMeltdown, effectstest.EaElectrosiphon, effectstest.EaVoyager, effectstest.EaGlassblower, effectstest.EaGlimmer,
		effectstest.EaHighspire, effectstest.EaInventor, effectstest.EaLiveFast, effectstest.EaReservoir, effectstest.EaRogueRefiner, effectstest.EaSage, effectstest.EaTune, effectstest.EaWoodweaver, effectstest.EaMaximus} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", oracle)
			continue
		}
		if spec.Completeness != effects.CompletenessFull {
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
		{"Sage of Shaila's Claim", effectstest.EaSage, "Creature — Elf Druid", 3},
		{"Reservoir Walker", effectstest.EaReservoir, "Artifact Creature — Construct", 3},
		{"Rogue Refiner", effectstest.EaRogueRefiner, "Creature — Human Rogue", 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			life, hand := me.Life, me.Hand.Size()
			effectstest.CastAndResolveCreature(t, g, tc.name, tc.typeLine, tc.oracle)
			effectstest.PassPriorityAroundTable(t, g)
			if got := effectstest.EnergyOf(me); got != tc.energy {
				t.Errorf("energy = %d, want %d", got, tc.energy)
			}
			switch tc.oracle {
			case effectstest.EaReservoir:
				if me.Life != life+3 {
					t.Errorf("life = %d, want %d", me.Life, life+3)
				}
			case effectstest.EaRogueRefiner:
				if me.Hand.Size() != hand+1 {
					t.Errorf("hand = %d, want %d (one drawn)", me.Hand.Size(), hand+1)
				}
			}
		})
	}
}

func TestEnergyBatchAAttuneWithAether(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedSearchLibrary(me, effectstest.SearchTestLand(effectstest.EaBasicLandMatch, "Basic Land — Forest"), effectstest.SearchTestLand("Odd", "Sorcery"))
	effectstest.CastCatalogSpell(t, g, "Attune with Aether", "Sorcery", effectstest.EaAttune, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.SearchChoiceFor(g, me.ID) != nil {
		effectstest.AnswerSearchNamed(t, g, me.ID, effectstest.EaBasicLandMatch)
		effectstest.PassPriorityAroundTable(t, g)
	}
	found := false
	for _, c := range me.Hand.Cards {
		if c.Name == effectstest.EaBasicLandMatch {
			found = true
		}
	}
	if !found {
		t.Error("the Forest is not in hand")
	}
	if effectstest.EnergyOf(me) != 2 {
		t.Errorf("energy = %d, want 2", effectstest.EnergyOf(me))
	}
}

func TestEnergyBatchAAetherMeltdownShrinksAndGivesEnergy(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := effectstest.PushVanillaCreature(g, opp.ID, "Victim", 5, 5)
	effectstest.CastCatalogSpell(t, g, "Aether Meltdown", "Enchantment — Aura", effectstest.EaMeltdown,
		[]game.TargetRef{{Kind: game.TargetCard, ID: victim}})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.EnergyOf(me) != 2 {
		t.Errorf("energy = %d, want 2", effectstest.EnergyOf(me))
	}
	if got := effectstest.ApaLive(g, victim).CurrentPower(); got != 1 {
		t.Errorf("enchanted creature's power = %d, want 1", got)
	}
}

func TestEnergyBatchAElectrosiphonCountersAndGivesItsManaValue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	spell := uuid.New()
	opp.Hand.PushTop(game.Card{InstanceID: spell, Name: "Big Instant", TypeLine: "Instant", ManaCost: "{3}{R}",
		Owner: opp.ID, Controller: opp.ID})
	if err := g.CastSpell(opp.ID, spell, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast the target: %v", err)
	}
	siphon := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: siphon, Name: "Electrosiphon", TypeLine: "Instant", OracleID: effectstest.EaElectrosiphon,
		Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, siphon, game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: spell}}}); err != nil {
		t.Fatalf("cast Electrosiphon: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(spell) {
		t.Error("the target was not countered")
	}
	if effectstest.EnergyOf(me) != 4 {
		t.Errorf("energy = %d, want 4 (its mana value)", effectstest.EnergyOf(me))
	}
}

func TestEnergyBatchAEmpyrealVoyagerGetsThatMany(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	v := effectstest.PushDiesCreatureForTest(g, me.ID, "Empyreal Voyager", effectstest.EaVoyager, "Creature — Vedalken Scout", 2, 3)
	effectstest.AttackWith(t, g, opp.ID, v)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.EnergyOf(me) != 2 {
		t.Errorf("energy = %d, want 2 (the damage dealt)", effectstest.EnergyOf(me))
	}
}

func TestEnergyBatchAGlassblowersPuzzleknot(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedLibrary(me, "A", "B", "C")
	knot := effectstest.CastAndResolveCreature(t, g, "Glassblower's Puzzleknot", "Artifact", effectstest.EaGlassblower)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.EnergyOf(me) != 0 {
		t.Fatalf("energy before the scry is answered = %d, want 0", effectstest.EnergyOf(me))
	}
	effectstest.AnswerScryKeepAll(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.EnergyOf(me) != 2 {
		t.Fatalf("energy after the enters scry = %d, want 2", effectstest.EnergyOf(me))
	}
	if err := g.ActivateCatalogAbility(me.ID, knot, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(knot) {
		t.Error("the Puzzleknot was not sacrificed")
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerScryKeepAll(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.EnergyOf(me) != 4 {
		t.Errorf("energy after the activation = %d, want 4", effectstest.EnergyOf(me))
	}
}

func TestEnergyBatchAWoodweaversPuzzleknot(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	life := me.Life
	knot := effectstest.CastAndResolveCreature(t, g, "Woodweaver's Puzzleknot", "Artifact", effectstest.EaWoodweaver)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.EnergyOf(me) != 3 || me.Life != life+3 {
		t.Fatalf("after entering: energy %d life %d, want 3 and %d", effectstest.EnergyOf(me), me.Life, life+3)
	}
	if err := g.ActivateCatalogAbility(me.ID, knot, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.EnergyOf(me) != 6 || me.Life != life+6 {
		t.Errorf("after the activation: energy %d life %d, want 6 and %d", effectstest.EnergyOf(me), me.Life, life+6)
	}
}

func TestEnergyBatchAGlimmerOfGenius(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedLibrary(me, "A", "B", "C", "D")
	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Glimmer of Genius", "Instant", effectstest.EaGlimmer, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerScryKeepAll(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+2 {
		t.Errorf("hand = %d, want %d", me.Hand.Size(), hand+2)
	}
	if effectstest.EnergyOf(me) != 2 {
		t.Errorf("energy = %d, want 2", effectstest.EnergyOf(me))
	}
}

func TestEnergyBatchAHighspireInfusion(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Highspire Infusion", "Instant", effectstest.EaHighspire, []game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.ApaLive(g, bear).CurrentPower(); got != 5 {
		t.Errorf("power = %d, want 5", got)
	}
	if effectstest.EnergyOf(me) != 2 {
		t.Errorf("energy = %d, want 2", effectstest.EnergyOf(me))
	}
}

func TestEnergyBatchAInspiredInventorModes(t *testing.T) {
	t.Run("energy", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		effectstest.CastAndResolveCreature(t, g, "Inspired Inventor", "Creature — Human Artificer", effectstest.EaInventor)
		c := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID)
		if c == nil {
			t.Fatal("no mode_pick prompt")
		}
		if err := g.ResolveModePick(c.ID, me.ID, []int{0}); err != nil {
			t.Fatal(err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if effectstest.EnergyOf(me) != 3 {
			t.Errorf("energy = %d, want 3", effectstest.EnergyOf(me))
		}
	})
	t.Run("counter", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
		effectstest.CastAndResolveCreature(t, g, "Inspired Inventor", "Creature — Human Artificer", effectstest.EaInventor)
		c := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID)
		if c == nil {
			t.Fatal("no mode_pick prompt")
		}
		if err := g.ResolveModePick(c.ID, me.ID, []int{1}); err != nil {
			t.Fatal(err)
		}
		effectstest.PickCard(t, g, me.ID, bear)
		effectstest.PassPriorityAroundTable(t, g)
		if n := effectstest.ApaLive(g, bear).Counters[game.CounterPlusOne]; n != 1 {
			t.Errorf("+1/+1 counters = %d, want 1", n)
		}
		if effectstest.EnergyOf(me) != 0 {
			t.Errorf("energy = %d, want 0", effectstest.EnergyOf(me))
		}
	})
}

func TestEnergyBatchALiveFastAndTuneTheNarrative(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	life, hand := me.Life, me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Live Fast", "Sorcery", effectstest.EaLiveFast, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+2 || me.Life != life-2 || effectstest.EnergyOf(me) != 2 {
		t.Errorf("Live Fast: hand %d life %d energy %d, want %d, %d and 2", me.Hand.Size(), me.Life, effectstest.EnergyOf(me), hand+2, life-2)
	}
	hand = me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Tune the Narrative", "Instant", effectstest.EaTune, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 || effectstest.EnergyOf(me) != 4 {
		t.Errorf("Tune the Narrative: hand %d energy %d, want %d and 4", me.Hand.Size(), effectstest.EnergyOf(me), hand+1)
	}
}

func TestEnergyBatchAMaximusSacrificesAnArtifactForEnergy(t *testing.T) {
	g, me, _ := effectstest.P7Table(t)
	maximus := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Maximus, Knight Apparent", OracleID: effectstest.EaMaximus,
		TypeLine: "Legendary Creature — Human Knight", Power: 4, Toughness: 4})
	if err := g.ActivateCatalogAbility(me.ID, maximus, 0, game.ActivateAbilityParams{}); err == nil {
		t.Fatal("activated with no artifact to sacrifice")
	}
	rock := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Rock", TypeLine: "Artifact"})
	effectstest.P7Activate(t, g, me, maximus, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{rock}})
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact was not sacrificed")
	}
	if effectstest.EnergyOf(me) != 2 {
		t.Errorf("energy = %d, want 2", effectstest.EnergyOf(me))
	}
	if e := effectstest.SpecActivatedEnergy(t, effectstest.EaMaximus, 0); e.Energy != 0 || e.Mana != "{1}" {
		t.Errorf("Maximus's cost = %+v, want {1} and a sacrifice", e)
	}
}

// Maximus's enters search is optional and finds only an Equipment card
// with mana value 2.
func TestEnergyBatchAMaximusSearchesForATwoDropEquipment(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Two Drop", TypeLine: "Artifact — Equipment", ManaCost: "{2}"},
		game.Card{Name: "Three Drop", TypeLine: "Artifact — Equipment", ManaCost: "{3}"})
	effectstest.CastAndResolveCreature(t, g, "Maximus, Knight Apparent", "Legendary Creature — Human Knight", effectstest.EaMaximus)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	if effectstest.SearchOptionNamed(g, c, "Three Drop") != uuid.Nil {
		t.Error("the search offered a mana value 3 Equipment")
	}
	effectstest.AnswerSearchNamed(t, g, me.ID, "Two Drop")
}
