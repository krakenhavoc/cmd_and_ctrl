package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// pr8Hit is one damage instruction (one instance): each source deals its
// amount to `to`, all at once, then the priority boundary runs.
func Pr8Hit(t *testing.T, g *game.Game, to uuid.UUID, hits map[uuid.UUID]int) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.DamageInstanceForEffect(func() error {
			for src, n := range hits {
				if err := g.DealMarkedDamageForEffect(src, nil, to, n, game.DamageMarks{}); err != nil {
					return err
				}
			}
			return nil
		}); err != nil {
			t.Fatal(err)
		}
	})
	g.RunStateChecksForTest()
}

// pr8Unpreventable deals `n` damage that can't be prevented (CR 615.12).
func Pr8Unpreventable(t *testing.T, g *game.Game, src, to uuid.UUID, n int) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.DealMarkedDamageForEffect(src, nil, to, n, game.DamageMarks{CantBePrevented: true}); err != nil {
			t.Fatal(err)
		}
	})
	g.RunStateChecksForTest()
}

// pr8Counters is the named counters on a battlefield permanent.
func Pr8Counters(g *game.Game, id uuid.UUID, kind string) int {
	c := FindBattlefieldCardForTest(g, id)
	if c == nil {
		return -1
	}
	return c.Counters[kind]
}

// pr8Enter puts a catalogued permanent onto the battlefield through the
// entry pipeline, so "enters with N counters" applies.
func Pr8Enter(t *testing.T, g *game.Game, owner *game.Player, name, typeLine, oracle string) uuid.UUID {
	t.Helper()
	c := game.NewCard(name, owner.ID)
	c.TypeLine, c.OracleID = typeLine, oracle
	owner.Hand.PushTop(c)
	var id uuid.UUID
	g.WithWriteLock(func() {
		var err error
		id, err = g.PutFromHandOntoBattlefieldForEffect(c.InstanceID, game.HandEntryOptions{Controller: owner.ID})
		if err != nil {
			t.Fatal(err)
		}
	})
	g.RunStateChecksForTest()
	return id
}
