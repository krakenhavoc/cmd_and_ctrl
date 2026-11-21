package testsr

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// revolt_cards_test.go — #2148: revolt and "a permanent you controlled
// left the battlefield this turn".

const (
	shortcutToMushroomsOracle    = "2dc6637a-0b51-462b-9f18-571146bb4558"
	fatalPushOracle              = "16437a83-be52-44cd-a768-a767c9347eb2"
	renegadeRallierOracle        = "6fa07b6c-f01a-4416-b0fc-986b0fc4e412"
	hiddenStockpileOracle        = "f5ded323-75c8-471e-a516-238b9d6b06d3"
	narnamRenegadeOracle         = "a4f99939-4b70-41ee-ad38-2dda56dbc9ce"
	silkweaverEliteOracle        = "3d0eb524-15e7-46f2-8f53-db08a07944d7"
	airdropAeronautsOracle       = "009a399b-c78e-475b-8bc3-7db2afc9d676"
	countlessGearsRenegadeOracle = "9d1f9036-a466-46b6-ab7c-2de487876978"
)

func TestNarnamRenegadeEntersWithACounterOnlyWithRevolt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	plain := effectstest.CastAndResolveCreature(t, g, "Narnam Renegade", "Creature — Elf Warrior", narnamRenegadeOracle)
	if n := effectstest.BattlefieldPlusOnes(g, plain); n != 0 {
		t.Errorf("no revolt: %d counters, want 0", n)
	}
	effectstest.RevoltFor(t, g, me.ID)
	boosted := effectstest.CastAndResolveCreature(t, g, "Narnam Renegade", "Creature — Elf Warrior", narnamRenegadeOracle)
	if n := effectstest.BattlefieldPlusOnes(g, boosted); n != 1 {
		t.Errorf("revolt: %d counters, want 1", n)
	}
}

func TestRevoltIgnoresAnOpponentsPermanentLeaving(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	effectstest.RevoltFor(t, g, opp.ID)
	id := effectstest.CastAndResolveCreature(t, g, "Narnam Renegade", "Creature — Elf Warrior", narnamRenegadeOracle)
	if n := effectstest.BattlefieldPlusOnes(g, id); n != 0 {
		t.Errorf("an opponent's permanent leaving gave %d counters, want 0", n)
	}
}

func TestSilkweaverEliteDrawsWithRevolt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	before := me.Hand.Size()
	effectstest.CastAndResolveCreature(t, g, "Silkweaver Elite", "Creature — Elf Archer", silkweaverEliteOracle)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != before {
		t.Fatalf("no revolt: hand %d → %d", before, me.Hand.Size())
	}
	effectstest.RevoltFor(t, g, me.ID)
	before = me.Hand.Size()
	effectstest.CastAndResolveCreature(t, g, "Silkweaver Elite", "Creature — Elf Archer", silkweaverEliteOracle)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != before+1 {
		t.Errorf("revolt: hand %d → %d, want +1", before, me.Hand.Size())
	}
}

func TestAirdropAeronautsGainsFiveWithRevolt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	life := me.Life
	effectstest.CastAndResolveCreature(t, g, "Airdrop Aeronauts", "Creature — Dwarf Scout", airdropAeronautsOracle)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life {
		t.Fatalf("no revolt: life %d → %d", life, me.Life)
	}
	effectstest.RevoltFor(t, g, me.ID)
	effectstest.CastAndResolveCreature(t, g, "Airdrop Aeronauts", "Creature — Dwarf Scout", airdropAeronautsOracle)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+5 {
		t.Errorf("revolt: life %d → %d, want +5", life, me.Life)
	}
}

func TestCountlessGearsRenegadeMakesAServoWithRevolt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.CastAndResolveCreature(t, g, "Countless Gears Renegade", "Creature — Dwarf Artificer", countlessGearsRenegadeOracle)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.OnBattlefieldNamed(g, "Servo"); n != 0 {
		t.Fatalf("no revolt made %d Servos", n)
	}
	// A TOKEN leaving counts too.
	g.WithWriteLock(func() {
		_ = g.CreateTokenForEffect(me.ID, effects.TokenCard("1/1 colorless Servo artifact"), 1)
	})
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Servo" {
			id := c.InstanceID
			g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(id) })
			break
		}
	}
	effectstest.CastAndResolveCreature(t, g, "Countless Gears Renegade", "Creature — Dwarf Artificer", countlessGearsRenegadeOracle)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.OnBattlefieldNamed(g, "Servo"); n != 1 {
		t.Errorf("revolt after a token died: %d Servos, want 1", n)
	}
}

func TestRenegadeRallierReturnsACheapPermanentWithRevolt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	cheap := effectstest.PushGraveyardPermanent(me, "Grizzly Bears", "Creature — Bear", "{1}{G}")
	dear := effectstest.PushGraveyardPermanent(me, "Serra Angel", "Creature — Angel", "{3}{W}{W}")

	effectstest.CastAndResolveCreature(t, g, "Renegade Rallier", "Creature — Human Warrior", renegadeRallierOracle)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Fatal("no revolt: the trigger asked for a target")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(cheap) {
		t.Fatal("no revolt: a card was returned")
	}

	effectstest.RevoltFor(t, g, me.ID)
	effectstest.CastAndResolveCreature(t, g, "Renegade Rallier", "Creature — Human Warrior", renegadeRallierOracle)
	prompt := effectstest.LatestPickTarget(g, me.ID)
	if prompt == nil {
		t.Fatal("revolt: no target prompt")
	}
	if !effectstest.HasID(prompt.PickTargetCards, cheap) || effectstest.HasID(prompt.PickTargetCards, dear) {
		t.Errorf("legal set %v should hold only the mana value 2 card", prompt.PickTargetCards)
	}
	effectstest.PickCard(t, g, me.ID, cheap)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(cheap) {
		t.Error("the chosen card did not return to the battlefield")
	}
}

func TestFatalPushMonitorsManaValueAndRevolt(t *testing.T) {
	cases := []struct {
		name   string
		cost   string
		revolt bool
		dies   bool
	}{
		{"mv2 no revolt", "{1}{G}", false, true},
		{"mv3 no revolt", "{2}{G}", false, false},
		{"mv4 revolt", "{3}{G}", true, true},
		{"mv5 revolt", "{4}{G}", true, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			victim := uuid.New()
			g.Battlefield.PushTop(game.Card{
				InstanceID: victim, Name: "Target", TypeLine: "Creature — Bear", ManaCost: tc.cost,
				Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID,
			})
			if tc.revolt {
				effectstest.RevoltFor(t, g, me.ID)
			}
			effectstest.CastCatalogSpell(t, g, "Fatal Push", "Instant", fatalPushOracle,
				[]game.TargetRef{{Kind: game.TargetCard, ID: victim}})
			effectstest.PassPriorityAroundTable(t, g)
			if died := !g.Battlefield.Contains(victim); died != tc.dies {
				t.Errorf("destroyed = %v, want %v", died, tc.dies)
			}
		})
	}
}

func TestShortcutToMushroomsCountersAtEndStepWithRevolt(t *testing.T) {
	// No revolt: nothing triggers.
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Shortcut to Mushrooms", "Enchantment", shortcutToMushroomsOracle, false)
	bear := effectstest.B12Creature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepEnd)
	if effectstest.TriggerOnStackFrom(g, "Shortcut to Mushrooms") || effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Fatal("triggered with no revolt")
	}

	// An opponent's permanent leaving is not revolt for me.
	g2 := effectstest.NewCatalogGame(t)
	me2, opp2 := g2.Seats[0], g2.Seats[1]
	effectstest.PushCatalogPermanent(g2, me2.ID, "Shortcut to Mushrooms", "Enchantment", shortcutToMushroomsOracle, false)
	effectstest.B12Creature(g2, me2.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	effectstest.RevoltFor(t, g2, opp2.ID)
	effectstest.AdvanceTo(t, g2, game.StepEnd)
	if effectstest.TriggerOnStackFrom(g2, "Shortcut to Mushrooms") || effectstest.LatestPickTarget(g2, me2.ID) != nil {
		t.Fatal("an opponent's permanent leaving triggered it")
	}

	// Revolt: pick a creature you control, it gets the counter.
	g3 := effectstest.NewCatalogGame(t)
	me3 := g3.Seats[0]
	effectstest.PushCatalogPermanent(g3, me3.ID, "Shortcut to Mushrooms", "Enchantment", shortcutToMushroomsOracle, false)
	bear3 := effectstest.B12Creature(g3, me3.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	effectstest.RevoltFor(t, g3, me3.ID)
	effectstest.AdvanceTo(t, g3, game.StepEnd)
	effectstest.PickCard(t, g3, me3.ID, bear3)
	effectstest.PassPriorityAroundTable(t, g3)
	if n := effectstest.BattlefieldPlusOnes(g3, bear3); n != 1 {
		t.Errorf("revolt: %d counters on the chosen creature, want 1", n)
	}
	_ = bear
}

func TestHiddenStockpileMakesAServoAtEndStepWithRevolt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Hidden Stockpile", "Enchantment", hiddenStockpileOracle, false)
	if len(game.ActivatedAbilitiesForCard(game.Card{OracleID: hiddenStockpileOracle})) != 1 {
		t.Error("the {1}, sacrifice a creature ability is not wired")
	}
	effectstest.RevoltFor(t, g, me.ID)
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.OnBattlefieldNamed(g, "Servo"); n != 1 {
		t.Errorf("revolt: %d Servos, want 1", n)
	}

	g2 := effectstest.NewCatalogGame(t)
	effectstest.PushCatalogPermanent(g2, g2.Seats[0].ID, "Hidden Stockpile", "Enchantment", hiddenStockpileOracle, false)
	effectstest.AdvanceTo(t, g2, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g2)
	if n := effectstest.OnBattlefieldNamed(g2, "Servo"); n != 0 {
		t.Errorf("no revolt: %d Servos, want 0", n)
	}
}
