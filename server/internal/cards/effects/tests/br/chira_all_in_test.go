package testsbr

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// chira_all_in_test.go — Chira, All In (#6101): one trigger for each
// instruction that discards, selects targets, chooses modes, flips coins or
// rolls dice at random, made by Chira's controller; nothing for another
// player's.

const (
	chiraOracle          = "f6552921-f949-4f3c-b75d-718c106d712b"
	chiraHellBentOracle  = "ee8ab29b-0749-463a-abbd-eb0b9013aec8"
	chiraGoblinPilot     = "4760bb87-56e9-47eb-b5ec-6fa4dd19156e"
	chiraMerrimentOracle = "78826624-8784-4003-8813-230f299a588f"
)

type chiraTable struct {
	g     *game.Game
	me    *game.Player
	opp   *game.Player
	chira uuid.UUID
}

func chiraGame(t *testing.T) chiraTable {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	chira := effectstest.B12Push(g, me.ID, "Chira, All In", "Legendary Creature — Human Rogue", chiraOracle, 3, 3)
	return chiraTable{g: g, me: me, opp: opp, chira: chira}
}

// payoffs is Chira's +1/+1 counters and its controller's tapped Treasures.
func (tb chiraTable) payoffs(t *testing.T) (counters, treasures int) {
	t.Helper()
	tb.g.ReadSnapshot(func() {
		for _, c := range tb.g.Battlefield.Cards {
			if c.InstanceID == tb.chira {
				counters = c.Counters[game.CounterPlusOne]
			}
			if c.Controller == tb.me.ID && c.IsToken() && c.Name == "Treasure" {
				if !c.Tapped {
					t.Errorf("the Treasure entered untapped")
				}
				treasures++
			}
		}
	})
	return counters, treasures
}

func (tb chiraTable) want(t *testing.T, what string, n int) {
	t.Helper()
	effectstest.PassPriorityAroundTable(t, tb.g)
	counters, treasures := tb.payoffs(t)
	if counters != n || treasures != n {
		t.Errorf("%s: Chira has %d counters and %d Treasures, want %d of each", what, counters, treasures, n)
	}
}

func TestChiraAllInDiscardAtRandomTriggersOncePerInstruction(t *testing.T) {
	tb := chiraGame(t)
	for i := 0; i < 3; i++ {
		effectstest.B20HandCard(tb.me, "Card", "Sorcery")
		effectstest.B20HandCard(tb.opp, "Card", "Sorcery")
	}
	tb.g.WithWriteLock(func() {
		if err := tb.g.DiscardRandomForEffect(tb.me.ID, 2); err != nil {
			t.Fatal(err)
		}
	})
	tb.want(t, "two cards discarded at random", 1)
	tb.g.WithWriteLock(func() {
		if err := tb.g.DiscardRandomForEffect(tb.opp.ID, 1); err != nil {
			t.Fatal(err)
		}
	})
	tb.want(t, "an opponent's random discard", 1)
}

// An empty hand discards nothing, so nothing was discarded at random.
func TestChiraAllInEmptyHandDiscardsNothing(t *testing.T) {
	tb := chiraGame(t)
	tb.g.WithWriteLock(func() {
		tb.me.Hand.Cards = nil
		if err := tb.g.DiscardRandomForEffect(tb.me.ID, 1); err != nil {
			t.Fatal(err)
		}
	})
	tb.want(t, "a random discard from an empty hand", 0)
}

// "Discard a card at random" as an activation cost (Hell-Bent Raider).
func TestChiraAllInRandomDiscardCost(t *testing.T) {
	tb := chiraGame(t)
	raider := effectstest.B12Push(tb.g, tb.me.ID, "Hell-Bent Raider", "Creature — Human Barbarian", chiraHellBentOracle, 2, 2)
	effectstest.B20HandCard(tb.me, "Card", "Sorcery")
	before := tb.me.Hand.Size()
	if err := tb.g.ActivateCatalogAbility(tb.me.ID, raider, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if tb.me.Hand.Size() != before-1 {
		t.Fatal("the cost discarded nothing")
	}
	tb.want(t, "a card discarded at random to pay a cost", 1)
}

// A target chosen at random (Goblin Test Pilot) is a target selected at
// random by the activator.
func TestChiraAllInTargetSelectedAtRandom(t *testing.T) {
	tb := chiraGame(t)
	pilot := effectstest.B12Push(tb.g, tb.me.ID, "Goblin Test Pilot", "Creature — Goblin Pilot Wizard", chiraGoblinPilot, 0, 2)
	if err := tb.g.ActivateCatalogAbility(tb.me.ID, pilot, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, tb.g)
	counters, treasures := tb.payoffs(t)
	if counters != 1 || treasures != 1 {
		// The Pilot may have drawn Chira itself as the target and dealt it
		// 2 damage: the counter is still there (a 4/4 with 2 damage).
		t.Errorf("a target selected at random: %d counters, %d Treasures, want 1 of each", counters, treasures)
	}
}

// "Choose one at random" (Outlaws' Merriment) is a mode chosen at random by
// the trigger's controller.
func TestChiraAllInModeChosenAtRandom(t *testing.T) {
	tb := chiraGame(t)
	effectstest.B12Push(tb.g, tb.me.ID, "Outlaws' Merriment", "Enchantment", chiraMerrimentOracle, 0, 0)
	effectstest.AdvanceToUpkeepOf(t, tb.g, 0)
	// Merriment's trigger and Chira's (the mode was drawn as Merriment's
	// went on the stack) are both waiting: order them, then resolve.
	// The Rogue bullet's enters trigger asks for a target: aim it at the
	// opponent whenever it is asked.
	for i := 0; i < 64; i++ {
		if p := effectstest.TriggerOrderPromptFor(tb.g, tb.me.ID); p != nil {
			if err := tb.g.ResolveTriggerOrder(p.ID, tb.me.ID, append([]uuid.UUID(nil), p.TriggerOrderIDs...)); err != nil {
				t.Fatalf("ResolveTriggerOrder: %v", err)
			}
			continue
		}
		if p := effectstest.LatestPickTarget(tb.g, tb.me.ID); p != nil {
			if err := tb.g.ResolvePickTarget(p.ID, tb.me.ID, effectstest.PlayerRef(tb.opp.ID)); err != nil {
				t.Fatalf("Rogue ping target: %v", err)
			}
			continue
		}
		if len(tb.g.StackMeta) == 0 && len(tb.g.PendingTriggers) == 0 && len(tb.g.PendingChoices) == 0 {
			break
		}
		if err := tb.g.PassPriority(); err != nil && !errors.Is(err, game.ErrChoicePending) && len(tb.g.PendingChoices) == 0 {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	modes := 0
	for _, ev := range tb.g.Events {
		if ev.Kind == game.EventModesChosenAtRandom {
			modes++
		}
	}
	if modes != 1 {
		t.Fatalf("%d random mode choices announced, want 1", modes)
	}
	counters, treasures := tb.payoffs(t)
	if counters != 1 || treasures != 1 {
		t.Errorf("a mode chosen at random: Chira has %d counters and %d Treasures, want 1 of each", counters, treasures)
	}
}

// Two coins in one instruction, then three dice in one instruction: one
// trigger each. Another player's flip triggers nothing.
func TestChiraAllInCoinsAndDiceTriggerOncePerInstruction(t *testing.T) {
	tb := chiraGame(t)
	tb.g.WithWriteLock(func() {
		if _, err := tb.g.FlipCoinsForEffect(game.RandomDraw{Player: tb.me.ID}, 2); err != nil {
			t.Fatal(err)
		}
	})
	tb.want(t, "two coins flipped", 1)
	tb.g.WithWriteLock(func() {
		if _, err := tb.g.RollDiceForEffect(game.RandomDraw{Player: tb.me.ID}, 6, 3); err != nil {
			t.Fatal(err)
		}
	})
	tb.want(t, "three dice rolled", 2)
	tb.g.WithWriteLock(func() {
		if _, err := tb.g.FlipCoinsForEffect(game.RandomDraw{Player: tb.opp.ID}, 1); err != nil {
			t.Fatal(err)
		}
	})
	tb.want(t, "an opponent's coin", 2)
}

// Grip of Chaos reselecting a spell's target at random is a target
// selection made by Grip's controller: Chira, under that player, triggers.
func TestChiraAllInTargetsReselectedAtRandom(t *testing.T) {
	tb := chiraGame(t)
	effectstest.PushCatalogPermanent(tb.g, tb.me.ID, "Grip of Chaos", "Enchantment", "ca2e0c8f-7ece-465f-9103-ba60712eacf7", false)
	a := effectstest.PushTypedCard(tb.g, tb.opp.ID, "Bear A", "Creature — Bear", "{1}{G}")
	effectstest.PushTypedCard(tb.g, tb.opp.ID, "Bear B", "Creature — Bear", "{1}{G}")
	effectstest.CastCatalogSpell(t, tb.g, "Doom Blade", "Instant", effectstest.DoomBladeOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: a}})
	// Chira is a legal target for the Blade too, so the draw may land on
	// it: count the Treasure, which survives either way.
	effectstest.PassPriorityAroundTable(t, tb.g)
	if _, treasures := tb.payoffs(t); treasures != 1 {
		t.Errorf("a spell's target reselected at random: %d Treasures, want 1", treasures)
	}
}
