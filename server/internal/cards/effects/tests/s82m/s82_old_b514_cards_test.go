package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b514_cards_test.go — Shards of Alara, batch 514 (S82 older sets,
// #5588). Every test plays the real catalog entry through the engine.

const (
	oldB514BloodCultistOracle = "031ddf1e-afcf-4b09-8859-4df0f0ab5dcf"
	oldB514GharialOracle      = "7020dc51-8271-40a3-9c4c-ebd40c0b08b2"
	oldB514DrakeOracle        = "7d51b860-f7a4-4b13-a83c-134931ae7baf"
	oldB514TarFiendOracle     = "82bc2ee1-9b51-4b3b-804b-1b4e8aab9a4f"
	oldB514HellionOracle      = "86a7dba7-aff9-4086-b149-03eee812249f"
	oldB514EmergenceOracle    = "e7f038af-977a-42ff-bebb-718b7490188a"
)

// oldB514Cast puts a creature card with printed stats in the active player's
// hand and casts it (the test pool pays no costs).
func oldB514Cast(t *testing.T, g *game.Game, name, oracle string, power, toughness int) uuid.UUID {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: "Creature — Test", OracleID: oracle,
		Power: power, Toughness: toughness, Owner: active.ID, Controller: active.ID,
	})
	effectstest.AdvanceToMain(t, g)
	if err := g.CastSpell(active.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell %s: %v", name, err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	return id
}

func TestOldB514EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB514BloodCultistOracle: "Blood Cultist",
		oldB514GharialOracle:      "Algae Gharial",
		oldB514DrakeOracle:        "Scavenger Drake",
		oldB514TarFiendOracle:     "Tar Fiend",
		oldB514HellionOracle:      "Caldera Hellion",
		oldB514EmergenceOracle:    "Mighty Emergence",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not in the catalog", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q completeness %v caveats %v, want Full with none", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func TestOldB514DrakeAndGharialGrowOnAnotherCreatureDyingIfYouSayYes(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle, keyword string
	}{
		{"Scavenger Drake", "Creature — Drake", oldB514DrakeOracle, "flying"},
		{"Algae Gharial", "Creature — Crocodile", oldB514GharialOracle, "shroud"},
	} {
		for _, yes := range []bool{true, false} {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			src := effectstest.B12Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, 1, 1)
			bear := effectstest.PushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
			effectstest.AdvanceToMain(t, g)
			if !effectstest.HasEffectiveKeyword(t, g, src, tc.keyword) {
				t.Errorf("%s lacks %s", tc.name, tc.keyword)
			}
			effectstest.KillCreature(t, g, opp.ID, bear)
			effectstest.PassPriorityAroundTable(t, g)
			if effectstest.OwedMays(g, me.ID) == 0 {
				t.Fatalf("%s: another creature dying must offer the counter", tc.name)
			}
			effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, yes)
			effectstest.PassPriorityAroundTable(t, g)
			want := 0
			if yes {
				want = 1
			}
			if got := effectstest.CounterCount(g, src, game.CounterPlusOne); got != want {
				t.Errorf("%s, answer %v: counters = %d, want %d", tc.name, yes, got, want)
			}
		}
	}
}

func TestOldB514DrakeDoesNotTriggerOnItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	src := effectstest.B12Push(g, me.ID, "Scavenger Drake", "Creature — Drake", oldB514DrakeOracle, 1, 1)
	effectstest.AdvanceToMain(t, g)
	effectstest.KillCreature(t, g, me.ID, src)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.OwedMays(g, me.ID) > 0 {
		t.Error("its own death offered a counter; the trigger says another creature")
	}
}

func TestOldB514BloodCultistPingsAndGrowsWhenTheVictimDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	cultist := effectstest.B12Push(g, me.ID, "Blood Cultist", "Creature — Human Wizard", oldB514BloodCultistOracle, 1, 1)
	survivor := effectstest.PushVanillaCreature(g, opp.ID, "Big Bear", 3, 3)
	runt := effectstest.PushVanillaCreature(g, opp.ID, "Runt", 1, 1)
	effectstest.AdvanceToMain(t, g)

	ping := func(target uuid.UUID) {
		t.Helper()
		if err := g.ActivateCatalogAbility(me.ID, cultist, 0, game.ActivateAbilityParams{
			Targets: []game.TargetRef{{Kind: game.TargetCard, ID: target}},
		}); err != nil {
			t.Fatalf("activate: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	ping(survivor)
	if got := effectstest.DamageOn(t, g, survivor); got != 1 {
		t.Errorf("damage on the survivor = %d, want 1", got)
	}
	if got := effectstest.CounterCount(g, cultist, game.CounterPlusOne); got != 0 {
		t.Errorf("counters after a ping that killed nothing = %d, want 0", got)
	}
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == cultist {
				g.Battlefield.Cards[i].Tapped = false
			}
		}
	})
	ping(runt)
	if got := effectstest.CounterCount(g, cultist, game.CounterPlusOne); got != 1 {
		t.Errorf("counters after the victim died = %d, want 1", got)
	}
	// A creature it did not damage dying does not count.
	other := effectstest.PushVanillaCreature(g, opp.ID, "Bystander", 1, 1)
	effectstest.KillCreature(t, g, opp.ID, other)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, cultist, game.CounterPlusOne); got != 1 {
		t.Errorf("counters after an undamaged creature died = %d, want still 1", got)
	}
}

func TestOldB514TarFiendMakesTheTargetDiscardPerCreatureDevoured(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := effectstest.PushDevourFodder(g, me.ID, "Goblin A")
	b := effectstest.PushDevourFodder(g, me.ID, "Goblin B")
	c1 := effectstest.PushCatalogHandCard(opp, "Spare 1", "Sorcery", "")
	c2 := effectstest.PushCatalogHandCard(opp, "Spare 2", "Sorcery", "")
	c3 := effectstest.PushCatalogHandCard(opp, "Spare 3", "Sorcery", "")
	id := effectstest.CastDevourer(t, g, "Tar Fiend", oldB514TarFiendOracle)
	effectstest.AnswerEntryChoice(t, g, game.PendingChoiceEntrySacrifice, me.ID, a, b)
	if got := effectstest.DevourPlusOnes(g, id); got != 4 {
		t.Errorf("counters = %d, want 4 (devour 2, two creatures)", got)
	}
	effectstest.AnswerPickTargetPlayer(t, g, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerDiscard(t, g, opp.ID, c1, c2)
	effectstest.PassPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(c1) || !opp.Graveyard.Contains(c2) || opp.Graveyard.Contains(c3) || !opp.Hand.Contains(c3) {
		t.Error("the target did not discard exactly the two cards they chose")
	}
}

func TestOldB514TarFiendDevouringNothingDiscardsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	spare := effectstest.PushCatalogHandCard(opp, "Spare", "Sorcery", "")
	id := effectstest.CastDevourer(t, g, "Tar Fiend", oldB514TarFiendOracle) // nothing to sacrifice: no prompt
	if got := effectstest.DevourPlusOnes(g, id); got != 0 {
		t.Errorf("counters = %d, want 0", got)
	}
	effectstest.AnswerPickTargetPlayerIfOpen(t, g, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if !opp.Hand.Contains(spare) {
		t.Error("the target discarded with nothing devoured")
	}
}

func TestOldB514CalderaHellionDamagesEveryCreatureIncludingItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	small := effectstest.PushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	big := effectstest.PushVanillaCreature(g, opp.ID, "Their Wall", 1, 4)
	mine := effectstest.PushVanillaCreature(g, me.ID, "My Bear", 2, 3)
	hellion := oldB514Cast(t, g, "Caldera Hellion", oldB514HellionOracle, 3, 3)
	effectstest.AnswerEntryChoice(t, g, game.PendingChoiceEntrySacrifice, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{small, mine, hellion} {
		if _, ok := effectstest.BattlefieldCard(g, id); ok {
			t.Errorf("creature %s survived 3 damage", id)
		}
	}
	if got := effectstest.DamageOn(t, g, big); got != 3 {
		t.Errorf("damage on the 4-toughness creature = %d, want 3", got)
	}
}

func TestOldB514CalderaHellionDevouredCountersLetItSurvive(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	fodder := effectstest.PushDevourFodder(g, me.ID, "Goblin")
	hellion := oldB514Cast(t, g, "Caldera Hellion", oldB514HellionOracle, 3, 3)
	effectstest.AnswerEntryChoice(t, g, game.PendingChoiceEntrySacrifice, me.ID, fodder)
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.BattlefieldCard(g, hellion); !ok {
		t.Fatal("a 4/4 Hellion died to its own 3 damage")
	}
	if got := effectstest.DamageOn(t, g, hellion); got != 3 {
		t.Errorf("damage on the Hellion = %d, want 3", got)
	}
}

func TestOldB514MightyEmergenceCountersOnlyPowerFiveOrMore(t *testing.T) {
	for _, tc := range []struct {
		power int
		yes   bool
		want  int
	}{{5, true, 2}, {5, false, 0}, {4, true, 0}} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		effectstest.PushCatalogPermanent(g, me.ID, "Mighty Emergence", "Enchantment", oldB514EmergenceOracle, false)
		id := oldB514Cast(t, g, "Big Thing", "", tc.power, 5)
		if tc.power < 5 {
			if effectstest.OwedMays(g, me.ID) != 0 {
				t.Errorf("power %d offered counters", tc.power)
			}
			continue
		}
		if effectstest.OwedMays(g, me.ID) == 0 {
			t.Fatalf("power %d did not offer counters", tc.power)
		}
		effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, tc.yes)
		effectstest.PassPriorityAroundTable(t, g)
		if got := effectstest.CounterCount(g, id, game.CounterPlusOne); got != tc.want {
			t.Errorf("power %d, answer %v: counters = %d, want %d", tc.power, tc.yes, got, tc.want)
		}
	}
}

func TestOldB514MightyEmergenceIgnoresAnOpponentsCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, opp.ID, "Mighty Emergence", "Enchantment", oldB514EmergenceOracle, false)
	oldB514Cast(t, g, "Big Thing", "", 6, 6)
	if effectstest.OwedMays(g, opp.ID) != 0 || effectstest.OwedMays(g, me.ID) != 0 {
		t.Error("an opponent's Mighty Emergence offered counters for my creature")
	}
}
