package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// mount_copy_saddlers_test.go — #3744 (split out of #3548, ADR 0186). The
// saddle record is per object (CR 702.171c, 707.2): a Clone or a token copy
// of a saddled Mount copies the Mount's characteristics but not the fact
// that it was saddled, so it reads only the creatures that saddled IT.
// The record is not a copiable value, so nothing needed to be built; these
// tests pin it so a future change to the copy path cannot start leaking.
// Every Mount reader (Gitrog, Fortune, Calamity, Giant Beaver) goes
// through SaddlersOf, so that is the surface asserted.

var saddleMounts = []struct {
	name, oracle string
}{
	{"The Gitrog, Ravenous Ride", gitrogRavenousRideOracle},
	{"Fortune, Loyal Steed", fortuneLoyalSteedOracle},
	{"Calamity, Galloping Inferno", "eb37496e-374f-41b9-bd52-f52f37bdac31"},
	{"Giant Beaver", "2a4b42b3-6a95-41e1-9741-b35723d8da73"},
}

func saddlersOfMount(g *game.Game, mount uuid.UUID) []uuid.UUID {
	var out []uuid.UUID
	g.WithWriteLock(func() { out = g.SaddlersOf(mount) })
	return out
}

// tokenCopyOfMount makes a token copy of the Mount and returns its id.
func tokenCopyOfMount(t *testing.T, g *game.Game, me *game.Player, src uuid.UUID) uuid.UUID {
	t.Helper()
	before := map[uuid.UUID]bool{}
	for _, c := range g.Battlefield.Cards {
		before[c.InstanceID] = true
	}
	if err := (CreateTokenCopy{Controller: me.ID, Copy: src, N: 1}).Apply(NewContext(g, nil)); err != nil {
		t.Fatalf("CreateTokenCopy: %v", err)
	}
	for _, c := range g.Battlefield.Cards {
		if !before[c.InstanceID] {
			return c.InstanceID
		}
	}
	t.Fatal("no token copy was created")
	return uuid.Nil
}

func TestTokenCopyOfASaddledMountReadsOnlyItsOwnSaddlers(t *testing.T) {
	for _, m := range saddleMounts {
		t.Run(m.name, func(t *testing.T) {
			g := newCatalogGame(t)
			toMain(t, g)
			me := g.Seats[g.Turn.ActiveSeat]
			mount := pushMountForTest(g, me.ID, m.name, m.oracle, 3)
			first := pushCrewerForTest(g, me.ID, "First saddler", 10)
			second := pushCrewerForTest(g, me.ID, "Second saddler", 10)

			if err := saddleAbility(t, g, me.ID, mount, first); err != nil {
				t.Fatalf("saddle original: %v", err)
			}
			passPriorityAroundTable(t, g)
			if got := saddlersOfMount(g, mount); len(got) != 1 || got[0] != first {
				t.Fatalf("original SaddlersOf = %v, want [%s]", got, first)
			}

			cp := tokenCopyOfMount(t, g, me, mount)
			if c := saddleCard(t, g, cp); c.Saddled || len(c.SaddledBy) != 0 {
				t.Fatalf("the copy inherited the record: Saddled=%v SaddledBy=%+v", c.Saddled, c.SaddledBy)
			}
			if got := saddlersOfMount(g, cp); len(got) != 0 {
				t.Fatalf("copy SaddlersOf = %v before it was saddled, want none", got)
			}

			if err := saddleAbility(t, g, me.ID, cp, second); err != nil {
				t.Fatalf("saddle copy: %v", err)
			}
			passPriorityAroundTable(t, g)
			if got := saddlersOfMount(g, cp); len(got) != 1 || got[0] != second {
				t.Errorf("copy SaddlersOf = %v, want only [%s]", got, second)
			}
			if got := saddlersOfMount(g, mount); len(got) != 1 || got[0] != first {
				t.Errorf("original SaddlersOf = %v after the copy saddled, want only [%s]", got, first)
			}
		})
	}
}

// A Clone entering as a copy of a saddled Mount is the replacement-effect
// path rather than the token path; same answer.
func TestCloneOfASaddledMountReadsOnlyItsOwnSaddlers(t *testing.T) {
	g := newCatalogGame(t)
	toMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	mount := pushMountForTest(g, me.ID, "Giant Beaver", saddleMounts[3].oracle, 3)
	first := pushCrewerForTest(g, me.ID, "First saddler", 10)
	second := pushCrewerForTest(g, me.ID, "Second saddler", 10)
	if err := saddleAbility(t, g, me.ID, mount, first); err != nil {
		t.Fatalf("saddle original: %v", err)
	}
	passPriorityAroundTable(t, g)

	cloneID := castCatalogSpell(t, g, "Clone", "Creature — Shapeshifter", oracleClone, nil)
	resolveWithCopyChoice(t, g, mount)

	if c := copyBattlefieldCard(t, g, cloneID); c.Name != "Giant Beaver" || c.Saddled || len(c.SaddledBy) != 0 {
		t.Fatalf("clone = %q saddled=%v by=%+v, want an unsaddled Giant Beaver", c.Name, c.Saddled, c.SaddledBy)
	}
	if got := saddlersOfMount(g, cloneID); len(got) != 0 {
		t.Fatalf("clone SaddlersOf = %v before it was saddled, want none", got)
	}
	// Saddle taps the OTHER creatures, so the Clone's summoning sickness is moot.
	if err := saddleAbility(t, g, me.ID, cloneID, second); err != nil {
		t.Fatalf("saddle clone: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := saddlersOfMount(g, cloneID); len(got) != 1 || got[0] != second {
		t.Errorf("clone SaddlersOf = %v, want only [%s]", got, second)
	}
}
