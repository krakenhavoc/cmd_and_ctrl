package legal_test

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// damage_split_2956_test.go — #2956, ADR 0147. CanonicalDamageSplit is
// what the view ships as a damage-assignment prompt's pre-filled answer,
// and, when it covers lethal for every blocker, what the client sends
// for a player with "Auto-assign combat damage" on. It is the same
// split the enumerator offers the bots.

// damageSplit is CanonicalDamageSplit for g's one pending prompt.
func damageSplit(t *testing.T, g *game.Game) legal.DamageSplit {
	t.Helper()
	var (
		split legal.DamageSplit
		ok    bool
	)
	g.WithWriteLock(func() {
		for _, c := range g.PendingChoices {
			if c != nil && c.Kind == game.PendingChoiceDamageAssignment {
				split, ok = legal.CanonicalDamageSplit(g, c)
				return
			}
		}
	})
	if !ok {
		t.Fatal("no canonical split for the damage-assignment prompt")
	}
	return split
}

func shares(s legal.DamageSplit) map[uuid.UUID]int {
	out := map[uuid.UUID]int{}
	for _, a := range s.Assignments {
		out[a.BlockerID] = a.Amount
	}
	return out
}

func TestCanonicalSplitTramplesTheRestOverWhenItKillsEveryBlocker(t *testing.T) {
	// The report on #2956: Vivi Ornitier, 44 power with trample, blocked
	// by creatures with 4, 3 and 2 toughness. Lethal to each and 35 to
	// the player, not 0.
	g, me, ids := blockedCombat(t, 44, 44, []string{"trample"}, [][2]int{{5, 4}, {3, 3}, {2, 2}})
	s := damageSplit(t, g)
	got := shares(s)
	if got[ids[0]] != 4 || got[ids[1]] != 3 || got[ids[2]] != 2 || s.TrampleTo != 35 {
		t.Fatalf("split %v, trample %d; want 4/3/2 and 35 over", got, s.TrampleTo)
	}
	if !s.CoversLethal || !slices.Equal(s.Lethal, []int{4, 3, 2}) {
		t.Fatalf("covers %v, lethal %v; want covers and [4 3 2]", s.CoversLethal, s.Lethal)
	}
	// The enumerator offers the same answer, and the engine takes it.
	m, offered, trample := assignmentMove(t, g, me)
	if offered[ids[0]] != 4 || trample != 35 {
		t.Fatalf("the enumerator offers %v (trample %d), not the split", offered, trample)
	}
	dispatchAll(t, g, me, []legal.Move{m})
}

func TestCanonicalSplitPutsTheRestOnTheLastBlockerWithoutTrample(t *testing.T) {
	g, me, ids := blockedCombat(t, 10, 10, nil, [][2]int{{2, 2}, {3, 3}})
	s := damageSplit(t, g)
	got := shares(s)
	if got[ids[0]] != 2 || got[ids[1]] != 8 || s.TrampleTo != 0 || !s.CoversLethal {
		t.Fatalf("split %v, trample %d, covers %v; want 2 then 8 on the last blocker", got, s.TrampleTo, s.CoversLethal)
	}
	m, _, _ := assignmentMove(t, g, me)
	dispatchAll(t, g, me, []legal.Move{m})
}

func TestCanonicalSplitCountsDeathtouchAndMarkedDamage(t *testing.T) {
	// CR 702.2c with CR 702.19b: deathtouch makes 1 lethal, so a
	// 5-power deathtouch trampler blocked by three 4/4s covers lethal
	// for all of them (3, not 12) and puts 1 on each and 2 over.
	g, me, ids := blockedCombat(t, 5, 5, []string{"trample", "deathtouch"}, [][2]int{{4, 4}, {4, 4}, {4, 4}})
	s := damageSplit(t, g)
	got := shares(s)
	if !slices.Equal(s.Lethal, []int{1, 1, 1}) || s.TrampleTo != 2 || !s.CoversLethal ||
		got[ids[0]] != 1 || got[ids[1]] != 1 || got[ids[2]] != 1 {
		t.Fatalf("deathtouch: split %v, lethal %v, trample %d, covers %v; want 1/1/1 and 2 over",
			got, s.Lethal, s.TrampleTo, s.CoversLethal)
	}
	m, _, _ := assignmentMove(t, g, me)
	dispatchAll(t, g, me, []legal.Move{m})

	// CR 702.19c: damage already marked counts toward lethal. The 2/4
	// with 3 marked needs only 1, so 4 power covers both (unmarked it
	// would need 5) and 2 go over.
	g, me, ids = blockedCombat(t, 4, 4, []string{"trample"}, [][2]int{{2, 4}, {1, 1}})
	if s = damageSplit(t, g); s.CoversLethal {
		t.Fatal("4 power claims to cover lethal 4 + 1 before any damage is marked")
	}
	g.WithWriteLock(func() { onBattlefield(g, ids[0]).DamageMarked = 3 })
	s = damageSplit(t, g)
	if !slices.Equal(s.Lethal, []int{1, 1}) || s.TrampleTo != 2 || !s.CoversLethal {
		t.Fatalf("marked damage: lethal %v, trample %d; want [1 1], 2 over", s.Lethal, s.TrampleTo)
	}
	m, _, _ = assignmentMove(t, g, me)
	dispatchAll(t, g, me, []legal.Move{m})
}

func TestCanonicalSplitShortOfLethalIsAChoice(t *testing.T) {
	g, _, ids := blockedCombat(t, 4, 4, []string{"trample"}, [][2]int{{1, 1}, {2, 4}})
	s := damageSplit(t, g)
	if s.CoversLethal {
		t.Fatal("4 power over lethal 1 + 4 claims to cover every blocker")
	}
	if got := shares(s); got[ids[1]] != 4 || s.TrampleTo != 0 {
		t.Fatalf("split %v (trample %d), want all 4 on the 2/4", got, s.TrampleTo)
	}
}
