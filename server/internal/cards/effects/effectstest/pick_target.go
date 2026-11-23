package effectstest

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// pick_target_test.go — S20 sub-PR 2: the ETB-destroy / recursion
// cards on real targets. Replaces the S19 auto-picker tests.

const (
	ReclamationSageOracle = "032ec6e2-6cc3-4a97-9cc7-3233f5e11904"
	AcidicSlimeOracle     = "21f45043-5419-4019-8b6c-e5294bd5f549"
	EternalWitnessOracle  = "30b24e8e-3b0e-4d8e-90f3-f66eb7c1858c"
)

func LatestPickTarget(g *game.Game, chooser uuid.UUID) *game.PendingChoice {
	for i := len(g.PendingChoices) - 1; i >= 0; i-- {
		c := g.PendingChoices[i]
		if c != nil && c.Kind == game.PendingChoicePickTarget && c.Chooser == chooser {
			return c
		}
	}
	return nil
}

func PickCard(t *testing.T, g *game.Game, chooser, cardID uuid.UUID) {
	t.Helper()
	p := LatestPickTarget(g, chooser)
	if p == nil {
		t.Fatalf("no pick_target prompt for %s", chooser)
	}
	if err := g.ResolvePickTarget(p.ID, chooser, game.TargetRef{Kind: game.TargetCard, ID: cardID}); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
}

func HasID(ids []uuid.UUID, id uuid.UUID) bool {
	for _, x := range ids {
		if x == id {
			return true
		}
	}
	return false
}

// castAndResolveCreature casts a catalog creature and resolves just
// the spell, leaving whatever it triggered pending.
func CastAndResolveCreature(t *testing.T, g *game.Game, name, typeLine, oracle string) uuid.UUID {
	t.Helper()
	id := CastCatalogSpell(t, g, name, typeLine, oracle, nil)
	for i := 0; i < 8 && g.Stack.Size() > 0; i++ {
		if err := g.PassPriority(); err != nil {
			// #730: a prompt queued on the way down (a cast trigger
			// that asks something) gates the table. Stop here — the
			// caller answers it and passes again.
			if errors.Is(err, game.ErrChoicePending) {
				break
			}
			t.Fatalf("PassPriority: %v", err)
		}
	}
	return id
}
