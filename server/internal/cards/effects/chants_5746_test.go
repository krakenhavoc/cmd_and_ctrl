package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// chants_5746_test.go — #5746: Thelon's Chant and Tourach's Chant, the
// Fallen Empires pay-unless whose payment is "put a -1/-1 counter on a
// creature you control" (game.PayActionMinusOneCounter).

const (
	thelonsChantOracle  = "ff4ccf8d-7d3d-4a44-a9ab-6bc88e90d6a7"
	tourachsChantOracle = "bdf4e225-23b5-47d9-9936-ddcec8aa1ca4"
)

func TestChantsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		thelonsChantOracle:  "Thelon's Chant",
		tourachsChantOracle: "Tourach's Chant",
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull {
			t.Errorf("%s: registered as %q, %v; want Full", oracle, spec.Name, spec.Completeness)
		}
		if len(spec.Triggered) != 2 {
			t.Errorf("%s has %d triggered abilities, want the upkeep and the land trigger", name, len(spec.Triggered))
		}
	}
}

// chantTable puts the Chant under seat 1 and returns the seat-0 player,
// who will play the land.
func chantTable(t *testing.T, name, oracle string) (*game.Game, *game.Player, *game.Player, uuid.UUID) {
	t.Helper()
	g := newCatalogGame(t)
	me, chanter := g.Seats[0], g.Seats[1]
	chant := pushPermanentForTest(g, chanter.ID, name, oracle, "Enchantment")
	return g, me, chanter, chant
}

// chantPrompt is the pay-unless the Chant's land trigger asks `chooser`,
// or nil.
func chantPrompt(g *game.Game, chooser uuid.UUID) *game.PendingChoice {
	for i := len(g.PendingChoices) - 1; i >= 0; i-- {
		c := g.PendingChoices[i]
		if c != nil && c.Kind == game.PendingChoicePayUnless && c.Chooser == chooser {
			return c
		}
	}
	return nil
}

func chantMinusCounters(t *testing.T, g *game.Game, id uuid.UUID) int {
	t.Helper()
	c, ok := battlefieldCard(g, id)
	if !ok {
		return -1
	}
	return c.Counters[game.CounterMinusOne]
}

// The player who puts the Swamp there is asked; the payment's options are
// the creatures they control, and paying puts the counter on the one they
// name instead of taking the damage.
func TestThelonsChantSwampPaidWithACounter(t *testing.T) {
	g, me, chanter, _ := chantTable(t, "Thelon's Chant", thelonsChantOracle)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	elf := b12Creature(g, me.ID, "Elf", "Creature — Elf", 1, 1)
	theirs := b12Creature(g, chanter.ID, "Their Bear", "Creature — Bear", 2, 2)
	b12Permanent(g, me.ID, "Rock", "Artifact")

	b13PlayAs(t, g, 0, "Swamp", "Basic Land — Swamp", "")
	passPriorityAroundTable(t, g)
	choice := chantPrompt(g, me.ID)
	if choice == nil {
		t.Fatalf("no pay-unless for the player who played the Swamp: %+v", g.PendingChoices)
	}
	a := choice.PayAction()
	if a == nil || a.Kind != game.PayActionMinusOneCounter || a.Count != 1 {
		t.Fatalf("the payment is %+v, want one -1/-1 counter", a)
	}
	var opts []uuid.UUID
	g.ReadSnapshot(func() { opts = g.PayActionOptionsForEffect(me.ID, a) })
	if len(opts) != 2 || !hasID(opts, bear) || !hasID(opts, elf) {
		t.Fatalf("the payment offers %v, want only the payer's two creatures", opts)
	}
	if err := g.ResolvePayUnlessWithCards(choice.ID, me.ID, true, []uuid.UUID{theirs}); err == nil {
		t.Fatal("another player's creature was accepted as the payment")
	}
	life := lifeOf(g, me.ID)
	if err := g.ResolvePayUnlessWithCards(choice.ID, me.ID, true, []uuid.UUID{bear}); err != nil {
		t.Fatalf("pay with the Bear: %v", err)
	}
	if got := chantMinusCounters(t, g, bear); got != 1 {
		t.Errorf("the Bear has %d -1/-1 counters, want 1", got)
	}
	if got := lifeOf(g, me.ID); got != life {
		t.Errorf("life %d -> %d although the counter was paid", life, got)
	}
	if got := chantMinusCounters(t, g, theirs); got != 0 {
		t.Errorf("the other player's creature got %d counters", got)
	}
}

// Declining is 3 damage from the Chant to that player; a player who
// controls no creature is not asked and takes it.
func TestThelonsChantSwampDeclinedOrUnpayableDealsThree(t *testing.T) {
	for _, tc := range []struct {
		name     string
		creature bool
	}{
		{"declined", true},
		{"no creature", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g, me, chanter, chant := chantTable(t, "Thelon's Chant", thelonsChantOracle)
			var bear uuid.UUID
			if tc.creature {
				bear = b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
			}
			// The chanter's creature is no payment for seat 0.
			b12Creature(g, chanter.ID, "Their Bear", "Creature — Bear", 2, 2)
			life := lifeOf(g, me.ID)
			b13PlayAs(t, g, 0, "Swamp", "Basic Land — Swamp", "")
			passPriorityAroundTable(t, g)
			choice := chantPrompt(g, me.ID)
			if tc.creature {
				if choice == nil {
					t.Fatal("a player with a creature was not asked")
				}
				if err := g.ResolvePayUnlessWithCards(choice.ID, me.ID, false, nil); err != nil {
					t.Fatalf("decline: %v", err)
				}
				if got := chantMinusCounters(t, g, bear); got != 0 {
					t.Errorf("a declined payment put %d counters on the Bear", got)
				}
			} else if choice != nil {
				t.Fatalf("a player with no creature was asked: %+v", choice)
			}
			if got := lifeOf(g, me.ID); got != life-3 {
				t.Errorf("life %d -> %d, want 3 damage", life, got)
			}
			if !g.Battlefield.Contains(chant) {
				t.Error("the Chant left the battlefield")
			}
		})
	}
}

// Thelon's Chant watches Swamps and Tourach's Chant Forests; neither sees
// the other's land.
func TestChantsWatchTheirOwnLandType(t *testing.T) {
	for _, tc := range []struct {
		name, oracle, land string
		hurts              bool
	}{
		{"Thelon's Chant", thelonsChantOracle, "Swamp", true},
		{"Thelon's Chant", thelonsChantOracle, "Forest", false},
		{"Tourach's Chant", tourachsChantOracle, "Forest", true},
		{"Tourach's Chant", tourachsChantOracle, "Swamp", false},
	} {
		t.Run(tc.name+" "+tc.land, func(t *testing.T) {
			g, me, _, _ := chantTable(t, tc.name, tc.oracle)
			life := lifeOf(g, me.ID)
			b13PlayAs(t, g, 0, tc.land, "Basic Land — "+tc.land, "")
			passPriorityAroundTable(t, g)
			want := life
			if tc.hurts {
				want = life - 3
			}
			if got := lifeOf(g, me.ID); got != want {
				t.Errorf("a %s: life %d -> %d, want %d", tc.land, life, got, want)
			}
		})
	}
}

// "That player" is the one who PUT the land there, even if it entered
// under someone else (the 2004-10-04 ruling): the event's PutBy, not the
// land's controller.
func TestTourachsChantAsksThePlayerWhoPutTheForest(t *testing.T) {
	g, me, chanter, chant := chantTable(t, "Tourach's Chant", tourachsChantOracle)
	other := g.Seats[2]
	forest := b12Permanent(g, other.ID, "Forest", "Basic Land — Forest")
	b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	b12Creature(g, other.ID, "Their Bear", "Creature — Bear", 2, 2)

	spec, _ := Lookup(tourachsChantOracle)
	item := &game.StackItem{Kind: game.StackItemTriggered, Controller: chanter.ID, Owner: chanter.ID, SourceCardID: chant,
		Trigger: &game.TriggerContext{Event: game.Event{Kind: game.EventETB, CardID: forest, PutBy: me.ID}}}
	var err error
	g.WithWriteLock(func() { err = spec.Triggered[1].Effect(g, item) })
	if err != nil {
		t.Fatalf("effect: %v", err)
	}
	if chantPrompt(g, other.ID) != nil {
		t.Error("the Forest's controller was asked, but it is not who put it there")
	}
	choice := chantPrompt(g, me.ID)
	if choice == nil {
		t.Fatal("the player who put the Forest there was not asked")
	}
	life, theirLife := lifeOf(g, me.ID), lifeOf(g, other.ID)
	if err := g.ResolvePayUnlessWithCards(choice.ID, me.ID, false, nil); err != nil {
		t.Fatalf("decline: %v", err)
	}
	if lifeOf(g, me.ID) != life-3 || lifeOf(g, other.ID) != theirLife {
		t.Errorf("life: putter %d -> %d, controller %d -> %d; want the putter dealt 3", life, lifeOf(g, me.ID), theirLife, lifeOf(g, other.ID))
	}
}

// The open question is a restore point: the restored table asks for the
// same payment, and both answers still work.
func TestTourachsChantPromptSurvivesARestore(t *testing.T) {
	for _, pay := range []bool{true, false} {
		g, me, _, _ := chantTable(t, "Tourach's Chant", tourachsChantOracle)
		bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
		b13PlayAs(t, g, 0, "Forest", "Basic Land — Forest", "")
		passPriorityAroundTable(t, g)
		if chantPrompt(g, me.ID) == nil {
			t.Fatal("no prompt before the restore")
		}
		life := lifeOf(g, me.ID)

		restored := restoreRoundTrip(t, g, true)
		choice := chantPrompt(restored, me.ID)
		if choice == nil {
			t.Fatal("the restore lost the prompt")
		}
		if a := choice.PayAction(); a == nil || a.Kind != game.PayActionMinusOneCounter {
			t.Fatalf("the restored payment is %+v", a)
		}
		var ids []uuid.UUID
		if pay {
			ids = []uuid.UUID{bear}
		}
		if err := restored.ResolvePayUnlessWithCards(choice.ID, me.ID, pay, ids); err != nil {
			t.Fatalf("answer after the restore (pay=%v): %v", pay, err)
		}
		wantCounters, wantLife := 0, life-3
		if pay {
			wantCounters, wantLife = 1, life
		}
		if got := chantMinusCounters(t, restored, bear); got != wantCounters {
			t.Errorf("pay=%v: the Bear has %d counters, want %d", pay, got, wantCounters)
		}
		if got := lifeOf(restored, me.ID); got != wantLife {
			t.Errorf("pay=%v: life %d, want %d", pay, got, wantLife)
		}
	}
}

// A 1/1 that takes the counter dies of it, and that is still a payment:
// no damage.
func TestThelonsChantCounterThatKillsStillPays(t *testing.T) {
	g, me, _, _ := chantTable(t, "Thelon's Chant", thelonsChantOracle)
	elf := b12Creature(g, me.ID, "Elf", "Creature — Elf", 1, 1)
	life := lifeOf(g, me.ID)
	b13PlayAs(t, g, 0, "Swamp", "Basic Land — Swamp", "")
	passPriorityAroundTable(t, g)
	choice := chantPrompt(g, me.ID)
	if choice == nil {
		t.Fatal("no prompt")
	}
	if err := g.ResolvePayUnlessWithCards(choice.ID, me.ID, true, []uuid.UUID{elf}); err != nil {
		t.Fatalf("pay with the Elf: %v", err)
	}
	if g.Battlefield.Contains(elf) {
		t.Error("a 1/1 with a -1/-1 counter survived")
	}
	if got := lifeOf(g, me.ID); got != life {
		t.Errorf("life %d -> %d although the payment was made", life, got)
	}
}

// The upkeep: pay {G} (Thelon's) or {B} (Tourach's), or sacrifice it.
func TestChantsUpkeepPayOrSacrifice(t *testing.T) {
	for _, tc := range []struct{ name, oracle, cost string }{
		{"Thelon's Chant", thelonsChantOracle, "{G}"},
		{"Tourach's Chant", tourachsChantOracle, "{B}"},
	} {
		for _, pay := range []bool{true, false} {
			g := newCatalogGame(t)
			me := g.Seats[0]
			chant := pushPermanentForTest(g, me.ID, tc.name, tc.oracle, "Enchantment")
			walkToUpkeepOfSeat0(t, g)
			passPriorityAroundTable(t, g)
			if pay {
				dcMana(t, g, me.ID, tc.cost)
			}
			answerPayUnless(t, g, me.ID, pay)
			if got := g.Battlefield.Contains(chant); got != pay {
				t.Errorf("%s, paid=%v: on the battlefield = %v", tc.name, pay, got)
			}
			if !pay && !me.Graveyard.Contains(chant) {
				t.Errorf("%s declined: not in its owner's graveyard", tc.name)
			}
		}
	}
}
