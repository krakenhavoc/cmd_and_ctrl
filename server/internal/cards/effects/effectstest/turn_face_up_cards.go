package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// manifestTop puts `c` on top of p's library and manifests it, returning
// the face-down permanent (controlled by p).
func ManifestTop(t *testing.T, g *game.Game, p *game.Player, c game.Card) uuid.UUID {
	t.Helper()
	c.InstanceID = uuid.New()
	c.Owner, c.Controller = p.ID, p.ID
	p.Library.PushTop(c)
	var id uuid.UUID
	g.WithWriteLock(func() {
		var err error
		if id, err = g.ManifestForEffect(p.ID); err != nil {
			t.Fatalf("ManifestForEffect: %v", err)
		}
	})
	if FindBattlefieldCardForTest(g, id) == nil {
		t.Fatal("nothing was manifested")
	}
	return id
}

func BearCard() game.Card {
	return game.Card{Name: "Sleeping Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Power: 2, Toughness: 2}
}

// answerOwnPermanents answers the open "choose from your permanents"
// prompt (ChoosePermanents on the chooser's own board).
func AnswerOwnPermanents(t *testing.T, g *game.Game, chooser uuid.UUID, picks ...uuid.UUID) {
	t.Helper()
	c := LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, chooser)
	if c == nil {
		t.Fatal("no own-permanents prompt is open")
	}
	if err := g.ResolveOwnPermanents(c.ID, chooser, picks); err != nil {
		t.Fatalf("ResolveOwnPermanents: %v", err)
	}
}
