package effects

import (
	"testing"

	"github.com/google/uuid"

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

// revoltFor makes a permanent `owner` controls leave the battlefield
// (a sacrificed land), the cheapest way to turn revolt on.
func revoltFor(t *testing.T, g *game.Game, owner uuid.UUID) {
	t.Helper()
	land := pushCatalogPermanent(g, owner, "Forest", "Basic Land — Forest", "", false)
	g.WithWriteLock(func() {
		if err := g.SacrificePermanentForEffect(land); err != nil {
			t.Fatalf("setup sacrifice: %v", err)
		}
	})
}

func battlefieldPlusOnes(g *game.Game, id uuid.UUID) int {
	return counterOnBattlefieldCard(g, id, game.CounterPlusOne)
}

func TestNarnamRenegadeEntersWithACounterOnlyWithRevolt(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	plain := castAndResolveCreature(t, g, "Narnam Renegade", "Creature — Elf Warrior", narnamRenegadeOracle)
	if n := battlefieldPlusOnes(g, plain); n != 0 {
		t.Errorf("no revolt: %d counters, want 0", n)
	}
	revoltFor(t, g, me.ID)
	boosted := castAndResolveCreature(t, g, "Narnam Renegade", "Creature — Elf Warrior", narnamRenegadeOracle)
	if n := battlefieldPlusOnes(g, boosted); n != 1 {
		t.Errorf("revolt: %d counters, want 1", n)
	}
}

func TestRevoltIgnoresAnOpponentsPermanentLeaving(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	revoltFor(t, g, opp.ID)
	id := castAndResolveCreature(t, g, "Narnam Renegade", "Creature — Elf Warrior", narnamRenegadeOracle)
	if n := battlefieldPlusOnes(g, id); n != 0 {
		t.Errorf("an opponent's permanent leaving gave %d counters, want 0", n)
	}
}

func TestSilkweaverEliteDrawsWithRevolt(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	before := me.Hand.Size()
	castAndResolveCreature(t, g, "Silkweaver Elite", "Creature — Elf Archer", silkweaverEliteOracle)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != before {
		t.Fatalf("no revolt: hand %d → %d", before, me.Hand.Size())
	}
	revoltFor(t, g, me.ID)
	before = me.Hand.Size()
	castAndResolveCreature(t, g, "Silkweaver Elite", "Creature — Elf Archer", silkweaverEliteOracle)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != before+1 {
		t.Errorf("revolt: hand %d → %d, want +1", before, me.Hand.Size())
	}
}

func TestAirdropAeronautsGainsFiveWithRevolt(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	life := me.Life
	castAndResolveCreature(t, g, "Airdrop Aeronauts", "Creature — Dwarf Scout", airdropAeronautsOracle)
	passPriorityAroundTable(t, g)
	if me.Life != life {
		t.Fatalf("no revolt: life %d → %d", life, me.Life)
	}
	revoltFor(t, g, me.ID)
	castAndResolveCreature(t, g, "Airdrop Aeronauts", "Creature — Dwarf Scout", airdropAeronautsOracle)
	passPriorityAroundTable(t, g)
	if me.Life != life+5 {
		t.Errorf("revolt: life %d → %d, want +5", life, me.Life)
	}
}

func TestCountlessGearsRenegadeMakesAServoWithRevolt(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	castAndResolveCreature(t, g, "Countless Gears Renegade", "Creature — Dwarf Artificer", countlessGearsRenegadeOracle)
	passPriorityAroundTable(t, g)
	if n := onBattlefieldNamed(g, "Servo"); n != 0 {
		t.Fatalf("no revolt made %d Servos", n)
	}
	// A TOKEN leaving counts too.
	g.WithWriteLock(func() {
		_ = g.CreateTokenForEffect(me.ID, TokenCard("1/1 colorless Servo artifact"), 1)
	})
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Servo" {
			id := c.InstanceID
			g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(id) })
			break
		}
	}
	castAndResolveCreature(t, g, "Countless Gears Renegade", "Creature — Dwarf Artificer", countlessGearsRenegadeOracle)
	passPriorityAroundTable(t, g)
	if n := onBattlefieldNamed(g, "Servo"); n != 1 {
		t.Errorf("revolt after a token died: %d Servos, want 1", n)
	}
}

func TestRenegadeRallierReturnsACheapPermanentWithRevolt(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	cheap := pushGraveyardPermanent(me, "Grizzly Bears", "Creature — Bear", "{1}{G}")
	dear := pushGraveyardPermanent(me, "Serra Angel", "Creature — Angel", "{3}{W}{W}")

	castAndResolveCreature(t, g, "Renegade Rallier", "Creature — Human Warrior", renegadeRallierOracle)
	if latestPickTarget(g, me.ID) != nil {
		t.Fatal("no revolt: the trigger asked for a target")
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(cheap) {
		t.Fatal("no revolt: a card was returned")
	}

	revoltFor(t, g, me.ID)
	castAndResolveCreature(t, g, "Renegade Rallier", "Creature — Human Warrior", renegadeRallierOracle)
	prompt := latestPickTarget(g, me.ID)
	if prompt == nil {
		t.Fatal("revolt: no target prompt")
	}
	if !hasID(prompt.PickTargetCards, cheap) || hasID(prompt.PickTargetCards, dear) {
		t.Errorf("legal set %v should hold only the mana value 2 card", prompt.PickTargetCards)
	}
	pickCard(t, g, me.ID, cheap)
	passPriorityAroundTable(t, g)
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
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			victim := uuid.New()
			g.Battlefield.PushTop(game.Card{
				InstanceID: victim, Name: "Target", TypeLine: "Creature — Bear", ManaCost: tc.cost,
				Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID,
			})
			if tc.revolt {
				revoltFor(t, g, me.ID)
			}
			castCatalogSpell(t, g, "Fatal Push", "Instant", fatalPushOracle,
				[]game.TargetRef{{Kind: game.TargetCard, ID: victim}})
			passPriorityAroundTable(t, g)
			if died := !g.Battlefield.Contains(victim); died != tc.dies {
				t.Errorf("destroyed = %v, want %v", died, tc.dies)
			}
		})
	}
}

func TestShortcutToMushroomsCountersAtEndStepWithRevolt(t *testing.T) {
	// No revolt: nothing triggers.
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Shortcut to Mushrooms", "Enchantment", shortcutToMushroomsOracle, false)
	bear := b12Creature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	advanceTo(t, g, game.StepEnd)
	if triggerOnStackFrom(g, "Shortcut to Mushrooms") || latestPickTarget(g, me.ID) != nil {
		t.Fatal("triggered with no revolt")
	}

	// An opponent's permanent leaving is not revolt for me.
	g2 := newCatalogGame(t)
	me2, opp2 := g2.Seats[0], g2.Seats[1]
	pushCatalogPermanent(g2, me2.ID, "Shortcut to Mushrooms", "Enchantment", shortcutToMushroomsOracle, false)
	b12Creature(g2, me2.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	revoltFor(t, g2, opp2.ID)
	advanceTo(t, g2, game.StepEnd)
	if triggerOnStackFrom(g2, "Shortcut to Mushrooms") || latestPickTarget(g2, me2.ID) != nil {
		t.Fatal("an opponent's permanent leaving triggered it")
	}

	// Revolt: pick a creature you control, it gets the counter.
	g3 := newCatalogGame(t)
	me3 := g3.Seats[0]
	pushCatalogPermanent(g3, me3.ID, "Shortcut to Mushrooms", "Enchantment", shortcutToMushroomsOracle, false)
	bear3 := b12Creature(g3, me3.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	revoltFor(t, g3, me3.ID)
	advanceTo(t, g3, game.StepEnd)
	pickCard(t, g3, me3.ID, bear3)
	passPriorityAroundTable(t, g3)
	if n := battlefieldPlusOnes(g3, bear3); n != 1 {
		t.Errorf("revolt: %d counters on the chosen creature, want 1", n)
	}
	_ = bear
}

func TestHiddenStockpileMakesAServoAtEndStepWithRevolt(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Hidden Stockpile", "Enchantment", hiddenStockpileOracle, false)
	if len(game.ActivatedAbilitiesForCard(game.Card{OracleID: hiddenStockpileOracle})) != 1 {
		t.Error("the {1}, sacrifice a creature ability is not wired")
	}
	revoltFor(t, g, me.ID)
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if n := onBattlefieldNamed(g, "Servo"); n != 1 {
		t.Errorf("revolt: %d Servos, want 1", n)
	}

	g2 := newCatalogGame(t)
	pushCatalogPermanent(g2, g2.Seats[0].ID, "Hidden Stockpile", "Enchantment", hiddenStockpileOracle, false)
	advanceTo(t, g2, game.StepEnd)
	passPriorityAroundTable(t, g2)
	if n := onBattlefieldNamed(g2, "Servo"); n != 0 {
		t.Errorf("no revolt: %d Servos, want 0", n)
	}
}
