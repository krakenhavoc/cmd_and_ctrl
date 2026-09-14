package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// mind_control_test.go pins the S24 layer-2 control change. The card
// is two lines; everything asserted here is the engine behind it.

const mindControlOracle = "5912546a-acc2-448c-b042-64bdac5ec129"

// controllerOf reads a battlefield card's controller through a
// snapshot, which forces the recompute — so it reads the
// post-layer-2 answer, which is the whole point of materialising.
func controllerOf(t *testing.T, g *game.Game, id uuid.UUID) uuid.UUID {
	t.Helper()
	var out uuid.UUID
	var found bool
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == id {
				out, found = c.Controller, true
				return
			}
		}
	})
	if !found {
		t.Fatalf("card %s not on battlefield", id)
	}
	return out
}

func summoningSickOf(t *testing.T, g *game.Game, id uuid.UUID) bool {
	t.Helper()
	var out bool
	g.ReadSnapshot(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				out = game.HasSummoningSickness(&g.Battlefield.Cards[i])
				return
			}
		}
	})
	return out
}

func TestMindControlStealsTheCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bear", TypeLine: testCreatureTypeLine,
		Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID,
	})
	if got := controllerOf(t, g, theirs); got != opp.ID {
		t.Fatalf("setup: controller %s, want %s", got, opp.ID)
	}

	aura := castCatalogSpell(t, g, "Mind Control", auraTypeLine, mindControlOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: theirs}})
	passPriorityAroundTable(t, g)

	if host := attachmentHostOf(t, g, aura); host.Kind != game.TargetCard || host.ID != theirs {
		t.Fatalf("Mind Control AttachedTo = %+v, want card %s", host, theirs)
	}
	if got := controllerOf(t, g, theirs); got != me.ID {
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
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bear", TypeLine: testCreatureTypeLine,
		Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID,
	})
	aura := castCatalogSpell(t, g, "Mind Control", auraTypeLine, mindControlOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: theirs}})
	passPriorityAroundTable(t, g)
	if got := controllerOf(t, g, theirs); got != me.ID {
		t.Fatalf("setup: controller %s, want %s", got, me.ID)
	}

	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(aura) })

	if got := controllerOf(t, g, theirs); got != opp.ID {
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
	g := newCatalogGame(t)
	_, opp := g.Seats[0], g.Seats[1]
	theirs := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bear", TypeLine: testCreatureTypeLine,
		Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID,
	})
	// pushBattlefieldCardWithTimestamp clears the flag: this creature
	// has been in play and is ready to act for its current controller.
	if summoningSickOf(t, g, theirs) {
		t.Fatal("setup: the creature should not be sick before the steal")
	}

	castCatalogSpell(t, g, "Mind Control", auraTypeLine, mindControlOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: theirs}})
	passPriorityAroundTable(t, g)

	if !summoningSickOf(t, g, theirs) {
		t.Error("a stolen creature is summoning-sick for its new controller (CR 302.6)")
	}
}

// CR 613.7 — two control-changing effects sort by timestamp and the
// later one wins. This falls out of the layer engine's existing sort
// with no code of its own, and is the thing a one-shot "set
// Controller" implementation gets wrong.
func TestSecondMindControlWinsOnTimestamp(t *testing.T) {
	g := newCatalogGame(t)
	first, second, victim := g.Seats[0], g.Seats[1], g.Seats[2]
	theirs := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bear", TypeLine: testCreatureTypeLine,
		Power: 2, Toughness: 2, Owner: victim.ID, Controller: victim.ID,
	})
	// Seat 0's copy resolves from its hand on its own turn.
	castCatalogSpell(t, g, "Mind Control", auraTypeLine, mindControlOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: theirs}})
	passPriorityAroundTable(t, g)
	if got := controllerOf(t, g, theirs); got != first.ID {
		t.Fatalf("first steal: controller %s, want %s", got, first.ID)
	}

	// Seat 1's copy arrives later, so it has the later CR 613.7d
	// attach timestamp.
	secondAura := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Mind Control", TypeLine: auraTypeLine,
		OracleID: mindControlOracle, Owner: second.ID, Controller: second.ID,
	})
	g.WithWriteLock(func() {
		if err := g.AttachForEffect(secondAura, game.TargetRef{Kind: game.TargetCard, ID: theirs}); err != nil {
			t.Fatalf("AttachForEffect: %v", err)
		}
	})
	if got := controllerOf(t, g, theirs); got != second.ID {
		t.Errorf("controller %s, want the later Aura's controller %s", got, second.ID)
	}

	// Remove the later one and the earlier one takes over again —
	// not the original owner.
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(secondAura) })
	if got := controllerOf(t, g, theirs); got != first.ID {
		t.Errorf("controller %s after the later Aura died, want %s", got, first.ID)
	}
}

// CR 704.5n composes with the control change: the Aura leaves
// because its host stopped being a legal creature, and control goes
// home in the same settling.
func TestMindControlFallsOffAndRevertsWhenTheHostLeaves(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bear", TypeLine: testCreatureTypeLine,
		Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID,
	})
	aura := castCatalogSpell(t, g, "Mind Control", auraTypeLine, mindControlOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: theirs}})
	passPriorityAroundTable(t, g)
	if got := controllerOf(t, g, theirs); got != me.ID {
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
		t.Error("CR 704.5n: the Aura should have gone to the graveyard")
	}
	if !me.Graveyard.Contains(aura) {
		t.Error("CR 704.5n: the Aura left the battlefield but did not reach its owner's graveyard")
	}
}
