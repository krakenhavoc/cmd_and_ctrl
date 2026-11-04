package effects

import (
	"errors"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// soc_b38_ancestral_communion_test.go: Ancestral Communion's "When you cast
// this spell while you control your commander" has no "if", so the commander
// is checked once, as the spell is cast (batch 38 fix, #4456).

func TestSocB38AncestralCommunionStillCopiesAfterTheCommanderLeaves(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	commander := msc9Commander(g, me.ID)
	a := batch01GraveyardCard(me, "Rock A", "Artifact")
	b := batch01GraveyardCard(me, "Rock B", "Artifact")
	castCatalogSpell(t, g, "Ancestral Communion", "Sorcery", msc9AncestralOracle, []game.TargetRef{{Kind: game.TargetCard, ID: a}})
	// The commander is gone before the cast trigger resolves. The wording has
	// no "if" (CR 603.4 does not apply), so the copy happens all the same.
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == commander {
				g.Battlefield.Cards[i].IsCommander = false
			}
		}
		if b18ControlsYourCommander(g, me.ID) {
			t.Fatal("the commander is still controlled")
		}
	})
	for i := 0; i < 60 && (g.Stack.Size() > 0 || len(g.PendingChoices) > 0); i++ {
		if ch := latestPickTarget(g, me.ID); ch != nil {
			if err := g.ResolvePickTarget(ch.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: b}); err != nil {
				t.Fatalf("new target for the copy: %v", err)
			}
			continue
		}
		if ch := latestConfirmFor(g, me.ID); ch != nil {
			if err := g.ResolveConfirm(ch.ID, me.ID, true); err != nil {
				t.Fatalf("confirm: %v", err)
			}
			continue
		}
		if err := g.PassPriority(); err != nil && !errors.Is(err, game.ErrChoicePending) {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if !me.Hand.Contains(a) || !me.Hand.Contains(b) {
		t.Errorf("the spell and its copy return both cards: a in hand %v, b in hand %v", me.Hand.Contains(a), me.Hand.Contains(b))
	}
}
