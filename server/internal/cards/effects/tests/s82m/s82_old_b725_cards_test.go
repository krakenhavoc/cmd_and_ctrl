package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b725_cards_test.go — Urza's Saga Auras, batch 725 (#5624). Each test
// plays the real catalog entry through the engine and asserts what its text says.

const (
	oldB725FieryMantle     = "1b553ed4-12eb-4a53-8e1f-d47ec92b1d65"
	oldB725ZephidsEmbrace  = "2f8b07ef-9d00-4eb6-a395-b5033aa3f80e"
	oldB725Fortitude       = "5017c07d-8ad5-4ee2-91d3-974811cdff75"
	oldB725Sicken          = "5208a5f2-eebe-4adc-8f29-543b60116817"
	oldB725PowerTaint      = "585a1c82-8a6c-4770-be9c-3a4916b74a9f"
	oldB725ParasiticBond   = "9d973df5-b2e8-4af8-8379-0f8282e1e62a"
	oldB725Launch          = "9e3b373f-05f5-44df-9568-bd1f8039aa54"
	oldB725VenomousFangs   = "a1ee715e-bab5-4edc-89ba-610cf71bb048"
	oldB725GaeasEmbrace    = "b44e3f2e-3185-40d5-b50c-d3a1e473120c"
	oldB725VampiricEmbrace = "c7a50ece-4fe4-4e54-9319-b92a4144445d"
	oldB725PendrellFlux    = "fe902db2-3b1d-4266-88eb-307f878fc317"
)

var oldB725Cards = map[string]string{
	oldB725FieryMantle: "Fiery Mantle", oldB725ZephidsEmbrace: "Zephid's Embrace", oldB725Fortitude: "Fortitude",
	oldB725Sicken: "Sicken", oldB725PowerTaint: "Power Taint", oldB725ParasiticBond: "Parasitic Bond",
	oldB725Launch: "Launch", oldB725VenomousFangs: "Venomous Fangs", oldB725GaeasEmbrace: "Gaea's Embrace",
	oldB725VampiricEmbrace: "Vampiric Embrace", oldB725PendrellFlux: "Pendrell Flux",
}

func TestOldB725EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range oldB725Cards {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func oldB725Setup(t *testing.T) (*game.Game, *game.Player, *game.Player, uuid.UUID) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	return g, me, opp, bear
}

func oldB725InHand(p *game.Player, name string) bool {
	for _, c := range p.Hand.Cards {
		if c.Name == name {
			return true
		}
	}
	return false
}

func TestOldB725FieryMantlePumpsAndReturnsToHand(t *testing.T) {
	g, me, _, bear := oldB725Setup(t)
	aura := effectstest.Enchant(t, g, "Fiery Mantle", oldB725FieryMantle, bear)
	effectstest.FloatMana(t, g, me, "{R}")
	effectstest.B16Activate(t, g, me.ID, aura, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if p, th := effectstest.EffectivePTOf(t, g, bear); p != 3 || th != 2 {
		t.Errorf("after {R} the creature is %d/%d, want 3/2", p, th)
	}
	effectstest.KillCreature(t, g, me.ID, bear)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if !oldB725InHand(me, "Fiery Mantle") {
		t.Error("the Aura did not return to its owner's hand")
	}
}

func TestOldB725ZephidsEmbraceGivesPlusTwoFlyingAndShroud(t *testing.T) {
	g, _, _, bear := oldB725Setup(t)
	effectstest.Enchant(t, g, "Zephid's Embrace", oldB725ZephidsEmbrace, bear)
	if p, th := effectstest.EffectivePTOf(t, g, bear); p != 4 || th != 4 {
		t.Errorf("the creature is %d/%d, want 4/4", p, th)
	}
	for _, kw := range []string{"flying", "shroud"} {
		if !effectstest.HasEffectiveKeyword(t, g, bear, kw) {
			t.Errorf("the creature lacks %s", kw)
		}
	}
}

func TestOldB725FortitudeSacrificesAForestToRegenerateAndReturns(t *testing.T) {
	g, me, opp, bear := oldB725Setup(t)
	forest := effectstest.PushLandFor(g, me.ID, "Forest", "Basic Land — Forest")
	aura := effectstest.Enchant(t, g, "Fortitude", oldB725Fortitude, bear)
	effectstest.B16Activate(t, g, me.ID, aura, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{forest}})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(forest) {
		t.Fatal("the Forest was not sacrificed")
	}
	src := effectstest.B12Creature(g, opp.ID, "Source", "Creature — Giant", 9, 9)
	effectstest.DamageAndSettle(t, g, src, bear, 5)
	if !g.Battlefield.Contains(bear) {
		t.Error("the regeneration shield did not save the creature")
	}
	effectstest.KillCreature(t, g, me.ID, bear)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if !oldB725InHand(me, "Fortitude") {
		t.Error("the Aura did not return to its owner's hand")
	}
}

func TestOldB725FortitudeNeedsAForest(t *testing.T) {
	g, me, _, bear := oldB725Setup(t)
	effectstest.PushLandFor(g, me.ID, "Island", "Basic Land — Island")
	aura := effectstest.Enchant(t, g, "Fortitude", oldB725Fortitude, bear)
	if err := g.ActivateCatalogAbility(me.ID, aura, 0, game.ActivateAbilityParams{}); err == nil {
		t.Error("activated Fortitude with no Forest to sacrifice")
	}
}

func TestOldB725SickenShrinksAndCycles(t *testing.T) {
	g, _, _, bear := oldB725Setup(t)
	effectstest.Enchant(t, g, "Sicken", oldB725Sicken, bear)
	if p, th := effectstest.EffectivePTOf(t, g, bear); p != 1 || th != 1 {
		t.Errorf("the creature is %d/%d, want 1/1", p, th)
	}
	g2 := effectstest.NewCatalogGame(t)
	id, me := effectstest.CycleFromHand(t, g2, "Sicken", effectstest.AuraTypeLine, oldB725Sicken, "{C}{C}")
	if !me.Graveyard.Contains(id) {
		t.Fatal("the cycled Sicken is not in the graveyard")
	}
	before := len(me.Hand.Cards)
	effectstest.PassPriorityAroundTable(t, g2)
	if len(me.Hand.Cards) != before+1 {
		t.Errorf("hand %d -> %d, want the cycling draw", before, len(me.Hand.Cards))
	}
}

func TestOldB725PowerTaintCostsTwoLifeUnlessPaidAtTheEnchantmentControllersUpkeep(t *testing.T) {
	for _, pay := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		ench := effectstest.TmpB1Push(g, opp.ID, "Their Enchantment", "Enchantment", "", 0, 0)
		effectstest.Enchant(t, g, "Power Taint", oldB725PowerTaint, ench)
		life, mine := opp.Life, me.Life
		effectstest.AdvanceToUpkeepOf(t, g, 1)
		if pay {
			effectstest.FloatMana(t, g, opp, "{C}{C}")
		}
		for i := 0; i < 8 && !effectstest.HasPayUnlessFor(g, opp.ID); i++ {
			if err := g.PassPriority(); err != nil {
				t.Fatalf("PassPriority: %v", err)
			}
		}
		effectstest.AnswerPayUnless(t, g, opp.ID, pay)
		effectstest.PassPriorityAroundTable(t, g)
		want := life - 2
		if pay {
			want = life
		}
		if me.Life != mine {
			t.Errorf("pay=%v: the Aura's controller lost life", pay)
		}
		if me.Life != mine {
			t.Errorf("pay=%v: the Aura's controller lost life", pay)
		}
		if opp.Life != want {
			t.Errorf("pay=%v: the enchantment's controller is at %d, want %d", pay, opp.Life, want)
		}
	}
}

func TestOldB725PowerTaintCannotEnchantACreature(t *testing.T) {
	g, me, _, bear := oldB725Setup(t)
	if err := effectstest.CastCatalogSpellErr(t, g, "Power Taint", effectstest.AuraTypeLine, oldB725PowerTaint,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}}); err == nil {
		t.Error("Power Taint targeted a creature")
	}
	_ = me
}

func TestOldB725PowerTaintCycles(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	id, me := effectstest.CycleFromHand(t, g, "Power Taint", effectstest.AuraTypeLine, oldB725PowerTaint, "{C}{C}")
	if !me.Graveyard.Contains(id) {
		t.Error("the cycled Power Taint is not in the graveyard")
	}
}

func TestOldB725ParasiticBondDamagesTheCreaturesControllerAtTheirUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	effectstest.Enchant(t, g, "Parasitic Bond", oldB725ParasiticBond, theirs)
	life, mine := opp.Life, me.Life
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-2 {
		t.Errorf("the creature's controller is at %d, want %d", opp.Life, life-2)
	}
	if me.Life != mine {
		t.Errorf("the Aura's controller lost life (%d -> %d)", mine, me.Life)
	}
}

func TestOldB725LaunchGivesFlyingAndReturnsToHand(t *testing.T) {
	g, me, _, bear := oldB725Setup(t)
	effectstest.Enchant(t, g, "Launch", oldB725Launch, bear)
	if !effectstest.HasEffectiveKeyword(t, g, bear, "flying") {
		t.Error("the creature lacks flying")
	}
	effectstest.KillCreature(t, g, me.ID, bear)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if !oldB725InHand(me, "Launch") {
		t.Error("the Aura did not return to its owner's hand")
	}
}

func TestOldB725VenomousFangsDestroysACreatureTheEnchantedOneDamaged(t *testing.T) {
	g, me, opp, bear := oldB725Setup(t)
	victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Elf", 1, 9)
	bystander := effectstest.B12Creature(g, opp.ID, "Bystander", "Creature — Elf", 1, 9)
	effectstest.Enchant(t, g, "Venomous Fangs", oldB725VenomousFangs, bear)
	// Damage a creature the enchanted one did not deal: nothing happens.
	other := effectstest.B12Creature(g, me.ID, "Other", "Creature — Bear", 2, 2)
	effectstest.DealDamageTo(t, g, other, bystander, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bystander) {
		t.Fatal("damage from an unenchanted creature destroyed the creature")
	}
	effectstest.DealDamageTo(t, g, bear, victim, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Error("the creature the enchanted one damaged was not destroyed")
	}
	if !g.Battlefield.Contains(bear) {
		t.Error("the enchanted creature was destroyed")
	}
}

func TestOldB725GaeasEmbracePumpsGrantsTrampleAndRegenerates(t *testing.T) {
	g, me, opp, bear := oldB725Setup(t)
	aura := effectstest.Enchant(t, g, "Gaea's Embrace", oldB725GaeasEmbrace, bear)
	if p, th := effectstest.EffectivePTOf(t, g, bear); p != 5 || th != 5 {
		t.Errorf("the creature is %d/%d, want 5/5", p, th)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "trample") {
		t.Error("the creature lacks trample")
	}
	effectstest.FloatMana(t, g, me, "{G}")
	effectstest.B16Activate(t, g, me.ID, aura, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	src := effectstest.B12Creature(g, opp.ID, "Source", "Creature — Giant", 9, 9)
	effectstest.DamageAndSettle(t, g, src, bear, 9)
	if !g.Battlefield.Contains(bear) {
		t.Error("the regeneration shield did not save the creature")
	}
}

func TestOldB725VampiricEmbracePumpsFliesAndGrowsWhenADamagedCreatureDies(t *testing.T) {
	g, me, opp, bear := oldB725Setup(t)
	victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Elf", 1, 3)
	bystander := effectstest.B12Creature(g, opp.ID, "Bystander", "Creature — Elf", 1, 1)
	effectstest.Enchant(t, g, "Vampiric Embrace", oldB725VampiricEmbrace, bear)
	if p, th := effectstest.EffectivePTOf(t, g, bear); p != 4 || th != 4 {
		t.Errorf("the creature is %d/%d, want 4/4", p, th)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "flying") {
		t.Error("the creature lacks flying")
	}
	effectstest.KillCreature(t, g, opp.ID, bystander)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CounterCount(g, bear, game.CounterPlusOne); n != 0 {
		t.Fatalf("a creature the enchanted one never damaged earned %d counters", n)
	}
	effectstest.DealDamageTo(t, g, bear, victim, 1)
	effectstest.KillCreature(t, g, opp.ID, victim)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CounterCount(g, bear, game.CounterPlusOne); n != 1 {
		t.Errorf("the enchanted creature has %d +1/+1 counters, want 1", n)
	}
	_ = me
}

func TestOldB725PendrellFluxSacrificesUnlessItsManaCostIsPaid(t *testing.T) {
	for _, pay := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		cre := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
			InstanceID: uuid.New(), Name: "Costly Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}",
			Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
		})
		effectstest.Enchant(t, g, "Pendrell Flux", oldB725PendrellFlux, cre)
		effectstest.AdvanceToUpkeepOf(t, g, 1)
		effectstest.PassPriorityAroundTable(t, g)
		if !g.Battlefield.Contains(cre) || effectstest.HasPayUnlessFor(g, me.ID) {
			t.Fatalf("pay=%v: the trigger fired at another player's upkeep", pay)
		}
		effectstest.AdvanceToUpkeepOf(t, g, 0)
		if pay {
			effectstest.FloatMana(t, g, me, "{C}{G}")
		}
		for i := 0; i < 8 && !effectstest.HasPayUnlessFor(g, me.ID); i++ {
			if err := g.PassPriority(); err != nil {
				t.Fatalf("PassPriority: %v", err)
			}
		}
		effectstest.AnswerPayUnless(t, g, me.ID, pay)
		effectstest.PassPriorityAroundTable(t, g)
		if g.Battlefield.Contains(cre) != pay {
			t.Errorf("pay=%v: creature on the battlefield = %v", pay, g.Battlefield.Contains(cre))
		}
	}
}
