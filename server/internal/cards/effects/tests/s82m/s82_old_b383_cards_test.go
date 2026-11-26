package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b383_cards_test.go — OLD batch 383 (Commander 2013, #5560): the five
// Curses and Flickerform. Each test plays the real catalog entry through the
// engine.

const (
	oldB383Inertia     = "0bbeb0ee-647b-43d3-91b3-6869d5ccb8b8"
	oldB383Predation   = "69ea4ce2-f749-4d5b-a392-efcfdce14a30"
	oldB383Chaos       = "a242af4b-5de6-4961-b684-f4bd809977c6"
	oldB383ShallowGrav = "dc66dbed-f979-4ad1-b5f9-08538886167f"
	oldB383Forsaken    = "e00f8d6b-5bb1-4625-9fa4-114cdc381bd4"
	oldB383Flickerform = "e5345c28-7046-4ff3-a5d6-eeb7a0fb230b"
)

func TestOldB383EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB383Inertia: "Curse of Inertia", oldB383Predation: "Curse of Predation", oldB383Chaos: "Curse of Chaos",
		oldB383ShallowGrav: "Curse of Shallow Graves", oldB383Forsaken: "Curse of the Forsaken", oldB383Flickerform: "Flickerform",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s (%s) is not registered", name, oracle)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name=%q completeness=%v caveats=%v, want Full with none", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func oldB383Curse(g *game.Game, owner, host *game.Player, name, oracle string) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: effectstest.CurseAuraTypeLine, OracleID: oracle,
		Owner: owner.ID, Controller: owner.ID,
		AttachedTo: game.TargetRef{Kind: game.TargetPlayer, ID: host.ID},
	})
}

func oldB383Bear(g *game.Game, owner *game.Player, name string) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: effectstest.TestCreatureTypeLine,
		Power: 2, Toughness: 2, Owner: owner.ID, Controller: owner.ID,
	})
}

// --- Curse of the Forsaken --------------------------------------------------

func TestOldB383ForsakenGivesEachAttackerItsControllerOneLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	oldB383Curse(g, third, opp, "Curse of the Forsaken", oldB383Forsaken)
	a, b := oldB383Bear(g, me, "Bear A"), oldB383Bear(g, me, "Bear B")
	life, curseOwner := me.Life, third.Life
	effectstest.DeclareAttack(t, g, opp.ID, a, b)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Life - life; got != 2 {
		t.Errorf("the attacker gained %d, want 2 (one per creature)", got)
	}
	if third.Life != curseOwner {
		t.Errorf("the Curse's controller's life changed %d → %d", curseOwner, third.Life)
	}
}

func TestOldB383ForsakenIgnoresAnAttackOnSomeoneElse(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	oldB383Curse(g, third, opp, "Curse of the Forsaken", oldB383Forsaken)
	a := oldB383Bear(g, me, "Bear A")
	life := me.Life
	effectstest.DeclareAttack(t, g, third.ID, a)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life {
		t.Errorf("life %d → %d for an attack on a player who isn't enchanted", life, me.Life)
	}
}

// --- Curse of Predation -----------------------------------------------------

func TestOldB383PredationPutsACounterOnEachAttacker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	oldB383Curse(g, opp, opp, "Curse of Predation", oldB383Predation)
	a, b := oldB383Bear(g, me, "Bear A"), oldB383Bear(g, me, "Bear B")
	effectstest.DeclareAttack(t, g, opp.ID, a, b)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b} {
		if n := effectstest.B12Counter(t, g, id, game.CounterPlusOne); n != 1 {
			t.Errorf("an attacker has %d +1/+1 counters, want 1", n)
		}
	}
}

// --- Curse of Chaos ---------------------------------------------------------

func TestOldB383ChaosAttackerMayDiscardThenDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	oldB383Curse(g, third, opp, "Curse of Chaos", oldB383Chaos)
	a, b := oldB383Bear(g, me, "Bear A"), oldB383Bear(g, me, "Bear B")
	hand := effectstest.HandSize(me)
	thirdHand := effectstest.HandSize(third)
	effectstest.DeclareAttack(t, g, opp.ID, a, b)
	effectstest.PassPriorityAroundTable(t, g)
	// One declaration, one trigger: one discard prompt for the attacker.
	if c := effectstest.ChooseCardsChoiceFor(g, me.ID); c == nil {
		t.Fatal("the attacking player was not offered a discard")
	}
	effectstest.SlChoose(t, g, me.ID, me.Hand.Cards[0].InstanceID)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.HandSize(me); got != hand {
		t.Errorf("hand %d → %d, want the same size after discarding and drawing", hand, got)
	}
	if me.Graveyard.Size() < 1 {
		t.Error("nothing was discarded")
	}
	if effectstest.HandSize(third) != thirdHand {
		t.Error("the Curse's controller's hand changed")
	}
	if c := effectstest.ChooseCardsChoiceFor(g, me.ID); c != nil {
		t.Error("two creatures from one declaration prompted twice")
	}
}

// --- Curse of Shallow Graves ------------------------------------------------

func TestOldB383ShallowGravesMakesATappedZombieForTheAttacker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	oldB383Curse(g, third, opp, "Curse of Shallow Graves", oldB383ShallowGrav)
	a := oldB383Bear(g, me, "Bear A")
	effectstest.DeclareAttack(t, g, opp.ID, a)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerMayChoice(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountOnBattlefield(g, "Zombie", me.ID); n != 1 {
		t.Fatalf("the attacker has %d Zombies, want 1", n)
	}
	if n := effectstest.CountOnBattlefield(g, "Zombie", third.ID); n != 0 {
		t.Errorf("the Curse's controller has %d Zombies, want 0", n)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Zombie" && !c.Tapped {
			t.Error("the Zombie entered untapped")
		}
	}
}

func TestOldB383ShallowGravesMayDecline(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	oldB383Curse(g, third, opp, "Curse of Shallow Graves", oldB383ShallowGrav)
	effectstest.DeclareAttack(t, g, opp.ID, oldB383Bear(g, me, "Bear A"))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerMayChoice(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountOnBattlefield(g, "Zombie", me.ID); n != 0 {
		t.Errorf("%d Zombies after declining", n)
	}
}

// --- Curse of Inertia -------------------------------------------------------

func TestOldB383InertiaLetsTheAttackerTapAPermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	oldB383Curse(g, third, opp, "Curse of Inertia", oldB383Inertia)
	target := oldB383Bear(g, opp, "Their Bear")
	effectstest.DeclareAttack(t, g, opp.ID, oldB383Bear(g, me, "Bear A"))
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, target)
	effectstest.PassPriorityAroundTable(t, g)
	// The choice between tapping and untapping is a prompt to the attacker.
	if effectstest.LatestOptionPickFor(g, me.ID) == nil {
		t.Fatal("the attacking player was not asked whether to tap or untap")
	}
	effectstest.AnswerOptionPick(t, g, me.ID, 0) // tap it
	effectstest.PassPriorityAroundTable(t, g)
	c, ok := effectstest.BattlefieldCard(g, target)
	if !ok || !c.Tapped {
		t.Error("the targeted permanent was not tapped")
	}
}

// --- Flickerform ------------------------------------------------------------

func TestOldB383FlickerformExilesTheCreatureAndItsAurasThenReturnsThemAttached(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	form := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Flickerform", TypeLine: "Enchantment — Aura", OracleID: oldB383Flickerform,
		Owner: me.ID, Controller: me.ID, AttachedTo: game.TargetRef{Kind: game.TargetCard, ID: bear},
	})
	other := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Pacifism", TypeLine: "Enchantment — Aura", Owner: opp.ID, Controller: opp.ID,
		AttachedTo: game.TargetRef{Kind: game.TargetCard, ID: bear},
	})
	effectstest.AddPool(t, g, me.ID, "{C}{C}{W}{W}")
	if err := g.ActivateCatalogAbility(me.ID, form, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{bear, form, other} {
		if z := effectstest.B12ZoneOf(g, id); z != game.ZoneExile {
			t.Fatalf("a card is in %s after the ability, want exile", z)
		}
	}
	effectstest.AdvanceToEndStepOf(t, g, g.Turn.ActiveSeat)
	effectstest.PassPriorityAroundTable(t, g)
	newBear := uuid.Nil
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Their Bear" {
			newBear = c.InstanceID
			if c.Controller != opp.ID {
				t.Errorf("the creature returned under %s, want its owner", c.Controller)
			}
		}
	}
	if newBear == uuid.Nil {
		t.Fatal("the creature did not return at the end step")
	}
	for _, name := range []string{"Flickerform", "Pacifism"} {
		found := false
		for _, c := range g.Battlefield.Cards {
			if c.Name != name {
				continue
			}
			found = true
			if c.AttachedTo.Kind != game.TargetCard || c.AttachedTo.ID != newBear {
				t.Errorf("%s is attached to %+v, want the returned creature", name, c.AttachedTo)
			}
		}
		if !found {
			t.Errorf("%s did not return", name)
		}
	}
}

func TestOldB383FlickerformCreatureWithNoAurasStillReturns(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	form := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Flickerform", TypeLine: "Enchantment — Aura", OracleID: oldB383Flickerform,
		Owner: me.ID, Controller: me.ID, AttachedTo: game.TargetRef{Kind: game.TargetCard, ID: bear},
	})
	effectstest.AddPool(t, g, me.ID, "{C}{C}{W}{W}")
	if err := g.ActivateCatalogAbility(me.ID, form, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Fatal("the creature is still on the battlefield")
	}
	effectstest.AdvanceToEndStepOf(t, g, g.Turn.ActiveSeat)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountOnBattlefield(g, "My Bear", me.ID); n != 1 {
		t.Errorf("%d Bears returned, want 1", n)
	}
}
