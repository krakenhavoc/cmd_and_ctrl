package testsm

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestMindControlStealsTheCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bear", TypeLine: effectstest.TestCreatureTypeLine,
		Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID,
	})
	if got := effectstest.ControllerOf(t, g, theirs); got != opp.ID {
		t.Fatalf("setup: controller %s, want %s", got, opp.ID)
	}

	aura := effectstest.CastCatalogSpell(t, g, "Mind Control", effectstest.AuraTypeLine, effectstest.MindControlOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: theirs}})
	effectstest.PassPriorityAroundTable(t, g)

	if host := effectstest.AttachmentHostOf(t, g, aura); host.Kind != game.TargetCard || host.ID != theirs {
		t.Fatalf("Mind Control AttachedTo = %+v, want card %s", host, theirs)
	}
	if got := effectstest.ControllerOf(t, g, theirs); got != me.ID {
		t.Errorf("controller %s, want the Aura's controller %s", got, me.ID)
	}
	// Ownership never changes (CR 108.3) — only control.
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == theirs && c.Owner != opp.ID {
				t.Errorf("owner changed to %s; control effects never touch ownership", c.Owner)
			}
		}
	})
}

// The reason this is a continuous effect and not a one-shot: nothing
// remembers the old controller, and control still goes home.
func TestMindControlRevertsWhenTheAuraLeaves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bear", TypeLine: effectstest.TestCreatureTypeLine,
		Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID,
	})
	aura := effectstest.CastCatalogSpell(t, g, "Mind Control", effectstest.AuraTypeLine, effectstest.MindControlOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: theirs}})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.ControllerOf(t, g, theirs); got != me.ID {
		t.Fatalf("setup: controller %s, want %s", got, me.ID)
	}

	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(aura) })

	if got := effectstest.ControllerOf(t, g, theirs); got != opp.ID {
		t.Errorf("controller %s after the Aura died, want %s", got, opp.ID)
	}
	if !g.Battlefield.Contains(theirs) {
		t.Error("the creature itself should be untouched")
	}
}

// CR 302.6 — a creature that changes control has summoning sickness
// under its new controller, so Mind Control does not hand over an
// attack on the turn it resolves.
func TestMindControlAppliesSummoningSickness(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	_, opp := g.Seats[0], g.Seats[1]
	theirs := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bear", TypeLine: effectstest.TestCreatureTypeLine,
		Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID,
	})
	// pushBattlefieldCardWithTimestamp clears the flag: this creature
	// has been in play and is ready to act for its current controller.
	if effectstest.SummoningSickOf(t, g, theirs) {
		t.Fatal("setup: the creature should not be sick before the steal")
	}

	effectstest.CastCatalogSpell(t, g, "Mind Control", effectstest.AuraTypeLine, effectstest.MindControlOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: theirs}})
	effectstest.PassPriorityAroundTable(t, g)

	if !effectstest.SummoningSickOf(t, g, theirs) {
		t.Error("a stolen creature is summoning-sick for its new controller (CR 302.6)")
	}
}

// CR 613.7 — two control-changing effects sort by timestamp and the
// later one wins. This falls out of the layer engine's existing sort
// with no code of its own, and is the thing a one-shot "set
// Controller" implementation gets wrong.
func TestSecondMindControlWinsOnTimestamp(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	first, second, victim := g.Seats[0], g.Seats[1], g.Seats[2]
	theirs := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bear", TypeLine: effectstest.TestCreatureTypeLine,
		Power: 2, Toughness: 2, Owner: victim.ID, Controller: victim.ID,
	})
	// Seat 0's copy resolves from its hand on its own turn.
	effectstest.CastCatalogSpell(t, g, "Mind Control", effectstest.AuraTypeLine, effectstest.MindControlOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: theirs}})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.ControllerOf(t, g, theirs); got != first.ID {
		t.Fatalf("first steal: controller %s, want %s", got, first.ID)
	}

	// Seat 1's copy arrives later, so it has the later CR 613.7e
	// attach timestamp.
	secondAura := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Mind Control", TypeLine: effectstest.AuraTypeLine,
		OracleID: effectstest.MindControlOracle, Owner: second.ID, Controller: second.ID,
	})
	g.WithWriteLock(func() {
		if err := g.AttachForEffect(secondAura, game.TargetRef{Kind: game.TargetCard, ID: theirs}); err != nil {
			t.Fatalf("AttachForEffect: %v", err)
		}
	})
	if got := effectstest.ControllerOf(t, g, theirs); got != second.ID {
		t.Errorf("controller %s, want the later Aura's controller %s", got, second.ID)
	}

	// Remove the later one and the earlier one takes over again —
	// not the original owner.
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(secondAura) })
	if got := effectstest.ControllerOf(t, g, theirs); got != first.ID {
		t.Errorf("controller %s after the later Aura died, want %s", got, first.ID)
	}
}

// CR 704.5m composes with the control change: the Aura leaves
// because its host stopped being a legal creature, and control goes
// home in the same settling.
func TestMindControlFallsOffAndRevertsWhenTheHostLeaves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bear", TypeLine: effectstest.TestCreatureTypeLine,
		Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID,
	})
	aura := effectstest.CastCatalogSpell(t, g, "Mind Control", effectstest.AuraTypeLine, effectstest.MindControlOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: theirs}})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.ControllerOf(t, g, theirs); got != me.ID {
		t.Fatalf("setup: controller %s, want %s", got, me.ID)
	}

	// The stolen creature dies. CR 400.3: to its OWNER's graveyard.
	if err := g.MoveCardByID(
		game.ZoneRef{Kind: game.ZoneBattlefield},
		game.ZoneRef{Kind: game.ZoneGraveyard, Owner: opp.ID},
		theirs,
	); err != nil {
		t.Fatalf("MoveCardByID: %v", err)
	}
	if g.Battlefield.Contains(aura) {
		t.Error("CR 704.5m: the Aura should have gone to the graveyard")
	}
	if !me.Graveyard.Contains(aura) {
		t.Error("CR 704.5m: the Aura left the battlefield but did not reach its owner's graveyard")
	}
}
