package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// werewolf_pack_leader_test.go — #4072: Werewolf Pack Leader (#4071).

const (
	hwWerewolfPackLeader = "086a911f-9ed4-4260-ae46-58c9333b95f5"
	hwMikaeus            = "5d27c63e-d1ef-48af-b51d-01ebc6daeac9"
)

func hwHasKeyword(c game.Card, kw string) bool {
	for _, a := range c.Effective().Abilities {
		if a == kw {
			return true
		}
	}
	return false
}

// hwDeclare declares the creatures as one attack declaration (CR 508.1),
// so a trigger that reads the whole declaration sees all of it.
func hwDeclare(t *testing.T, g *game.Game, defender uuid.UUID, attackers ...uuid.UUID) {
	t.Helper()
	advanceTo(t, g, game.StepDeclareAttackers)
	var decls []game.AttackDeclaration
	for _, id := range attackers {
		decls = append(decls, game.AttackDeclaration{Attacker: id, Target: defender})
	}
	if _, err := g.DeclareAttackers(decls); err != nil {
		t.Fatalf("DeclareAttackers: %v", err)
	}
}

func TestHWWerewolfPackLeaderIsFull(t *testing.T) {
	spec, ok := Lookup(hwWerewolfPackLeader)
	if !ok || spec.Completeness != CompletenessFull {
		t.Fatal("Werewolf Pack Leader is registered complete")
	}
}

func TestHWWerewolfPackLeaderPackTacticsDrawsOnlyWithSixPower(t *testing.T) {
	for _, tc := range []struct {
		name  string
		other int
		draws int
	}{{"total power 6", 3, 1}, {"total power 5", 2, 0}} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			leader := pushDiesCreatureForTest(g, me.ID, "Werewolf Pack Leader", hwWerewolfPackLeader, "Creature — Human Werewolf", 3, 3)
			other := pushVanillaCreature(g, me.ID, "Bear", tc.other, tc.other)
			before := me.Hand.Size()
			hwDeclare(t, g, opp.ID, leader, other)
			passPriorityAroundTable(t, g)
			if got := me.Hand.Size() - before; got != tc.draws {
				t.Errorf("cards drawn = %d, want %d", got, tc.draws)
			}
		})
	}
}

func TestHWWerewolfPackLeaderActivationSetsBaseStatsAndTrample(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	leader := pushDiesCreatureForTest(g, me.ID, "Werewolf Pack Leader", hwWerewolfPackLeader, "Creature — Human Werewolf", 3, 3)
	advanceToStepInTurn(t, g, game.StepPrecombatMain)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{G}{C}{C}{C}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, leader, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("ActivateCatalogAbility: %v", err)
	}
	passPriorityAroundTable(t, g)
	c := layeredCard(t, g, leader)
	if c.CurrentPower() != 5 || c.CurrentToughness() != 3 || !hwHasKeyword(c, "trample") {
		t.Errorf("leader %d/%d trample=%v, want 5/3 trample", c.CurrentPower(), c.CurrentToughness(), hwHasKeyword(c, "trample"))
	}
}

// While the ability lasts the Pack Leader isn't a Human, so Mikaeus'
// "other non-Human creatures" pump reaches it; at end of turn it is a
// Human again and does not (CR 611.2, 514.2).
func TestHWWerewolfPackLeaderStopsBeingAHumanUntilEndOfTurn(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	rfcbPush(g, me.ID, "Mikaeus, the Unhallowed", "Legendary Creature — Zombie Cleric", hwMikaeus, 3, 3)
	leader := pushDiesCreatureForTest(g, me.ID, "Werewolf Pack Leader", hwWerewolfPackLeader, "Creature — Human Werewolf", 3, 3)
	if c := layeredCard(t, g, leader); c.CurrentPower() != 3 || !c.HasSubtype("Human") {
		t.Fatalf("control: %d power, human %v, want a 3-power Human", c.CurrentPower(), c.HasSubtype("Human"))
	}
	advanceToStepInTurn(t, g, game.StepPrecombatMain)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{G}{C}{C}{C}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, leader, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("ActivateCatalogAbility: %v", err)
	}
	passPriorityAroundTable(t, g)
	c := layeredCard(t, g, leader)
	if c.HasSubtype("Human") || !c.HasSubtype("Werewolf") {
		t.Errorf("subtypes %v, want a Werewolf that is not a Human", c.Effective().Subtypes)
	}
	if c.CurrentPower() != 6 || c.CurrentToughness() != 4 || !hwHasKeyword(c, "trample") {
		t.Errorf("leader %d/%d trample=%v, want 6/4 (5/3 plus Mikaeus) trample", c.CurrentPower(), c.CurrentToughness(), hwHasKeyword(c, "trample"))
	}
	for i := 0; i < 60 && g.Turn.ActiveSeat == seat; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	c = layeredCard(t, g, leader)
	if !c.HasSubtype("Human") || c.CurrentPower() != 3 || hwHasKeyword(c, "trample") {
		t.Errorf("next turn: human %v, power %d, trample %v, want a 3-power Human without trample", c.HasSubtype("Human"), c.CurrentPower(), hwHasKeyword(c, "trample"))
	}
}
