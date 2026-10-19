package game

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
)

func TestScopedBlockColorReadsLiveColorsAndSurvivesRestore(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := pushCombatant(t, g, me, "Attacker", 2, 2)
	otherAttacker := pushCombatant(t, g, me, "Other attacker", 2, 2)
	blocker := pushCombatant(t, g, opp, "Blocker", 2, 2)
	g.WithWriteLock(func() {
		g.RegisterScopedEffectForEffect(uuid.Nil, g.PinnedObjectsLocked(attacker), []Mod{CantBeBlockedByColorMod("R")}, g.UntilEndOfTurnDuration(), "red evasion")
		g.RegisterScopedEffectForEffect(uuid.Nil, g.PinnedObjectsLocked(blocker), []Mod{SetColorsMod("R", "U")}, IndefiniteDuration(), "gold")
		g.RecomputeLayersIfStaleLocked()
	})
	check := func(g *Game, a, b uuid.UUID) BlockRefusal {
		var r BlockRefusal
		g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked(); r = g.BlockPairRefusalLocked(findCard(g, a), findCard(g, b)) })
		return r
	}
	if r := check(g, attacker, blocker); r.Reason != BlockReasonCantBeBlockedBy || r.Label != "red creatures" {
		t.Fatalf("multicolor blocker refusal: %+v", r)
	}
	if !check(g, otherAttacker, blocker).Legal() {
		t.Fatal("rule leaked to another attacker")
	}
	snap := g.CaptureSnapshot()
	if !snap.Restorable() {
		t.Fatalf("not restorable: %+v", snap.Continuations)
	}
	raw, err := json.Marshal(snap)
	if err != nil {
		t.Fatal(err)
	}
	var decoded GameSnapshot
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatal(err)
	}
	restored, err := decoded.RestoreStrict()
	if err != nil {
		t.Fatal(err)
	}
	if check(restored, attacker, blocker).Legal() {
		t.Fatal("restored effect lost its color")
	}
	restored.WithWriteLock(func() {
		restored.RegisterScopedEffectForEffect(uuid.Nil, restored.PinnedObjectsLocked(blocker), []Mod{SetColorsMod("U")}, IndefiniteDuration(), "blue")
	})
	if !check(restored, attacker, blocker).Legal() {
		t.Fatal("blocker remained barred after changing to blue")
	}
}

func TestScopedBlockColorValidatesStoredParameters(t *testing.T) {
	for _, m := range []Mod{
		{Kind: ModCantBeBlockedByColor, Text: "red creatures"},
		{Kind: ModCantBeBlockedByColor, Colors: []string{"C"}, Text: "colorless"},
		{Kind: ModCantBeBlockedByColor, Colors: []string{"R", "U"}, Text: "two colors"},
		{Kind: ModCantBeBlockedByColor, Colors: []string{"R"}},
	} {
		if blockRuleModProblem(m) == "" {
			t.Errorf("invalid rule accepted: %+v", m)
		}
	}
	if why := blockRuleModProblem(CantBeBlockedByColorMod("G")); why != "" {
		t.Fatal(why)
	}
}
