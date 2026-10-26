package game

import (
	"testing"

	"github.com/google/uuid"
)

// cantBeBlockedByObject (#4307, Wrinkly Monkey Shenanigans): the one named
// blocker, as the object it is (ID and epoch, CR 400.7), cannot block the
// pinned attacker; no other blocker is refused and no other attacker is
// protected from it.
func TestCantBeBlockedByObjectBarsOnlyThatBlockerFromThatAttacker(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := pushCombatant(t, g, me, "Attacker", 2, 2)
	other := pushCombatant(t, g, me, "Other Attacker", 2, 2)
	barred := pushCombatant(t, g, opp, "Barred Blocker", 2, 2)
	free := pushCombatant(t, g, opp, "Free Blocker", 2, 2)

	g.WithWriteLock(func() {
		b := findCard(g, barred)
		ref := ObjectRef{ID: b.InstanceID, Epoch: b.ObjectEpoch}
		if !g.RegisterScopedEffectForEffect(uuid.Nil, g.PinnedObjectsLocked(attacker),
			[]Mod{CantBeBlockedByObjectMod(ref, b.Name)}, IndefiniteDuration(), "test") {
			t.Fatal("setup: registered nothing")
		}
		g.RecomputeLayersIfStaleLocked()
	})
	refusal := func(att, blk uuid.UUID) BlockRefusal {
		var r BlockRefusal
		g.WithWriteLock(func() {
			r = g.blockRuleRefusalLocked(findCard(g, att), findCard(g, blk))
		})
		return r
	}
	if r := refusal(attacker, barred); r.Legal() {
		t.Error("the named blocker was allowed to block the attacker")
	} else if r.Reason != BlockReasonCantBeBlockedBy || r.Label != "Barred Blocker" {
		t.Errorf("refusal = %+v", r)
	}
	if r := refusal(attacker, free); !r.Legal() {
		t.Errorf("another blocker was refused: %+v", r)
	}
	if r := refusal(other, barred); !r.Legal() {
		t.Errorf("the named blocker was barred from another attacker: %+v", r)
	}
	// A new object (a new epoch) is not the one named (CR 400.7).
	g.WithWriteLock(func() { findCard(g, barred).ObjectEpoch++ })
	if r := refusal(attacker, barred); !r.Legal() {
		t.Errorf("a new object was still barred: %+v", r)
	}
}

func TestBlockRuleModProblemChecksTheObjectRule(t *testing.T) {
	if blockRuleModProblem(Mod{Kind: ModCantBeBlockedByObject, Text: "x"}) == "" {
		t.Error("a rule naming no blocker passed")
	}
	ref := ObjectRef{ID: uuid.New()}
	if blockRuleModProblem(Mod{Kind: ModCantBeBlockedByObject, Objects: []ObjectRef{ref}}) == "" {
		t.Error("a rule with no refusal text passed")
	}
	if p := blockRuleModProblem(CantBeBlockedByObjectMod(ref, "Barred Blocker")); p != "" {
		t.Errorf("unexpected problem %q", p)
	}
}
