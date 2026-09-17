package protocol

import (
	"strings"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// TestChooseCardsOverLookedAtLibraryCardsIsPrivate — #745. "Look at the
// top eight cards of your library. You may put a Dragon creature card
// from among them onto the battlefield" (Ureni) is the first
// choose_cards prompt whose candidates sit in a LIBRARY. The looker
// must be able to read them; the other seat must learn neither the
// cards, nor how many there are, nor any of their instance IDs.
func TestChooseCardsOverLookedAtLibraryCardsIsPrivate(t *testing.T) {
	g := newTwoSeatGame(t)
	me, them := g.Seats[0], g.Seats[1]
	g.WithWriteLock(func() {
		ids := g.LookAtTopOfLibraryForEffect(me.ID, 3)
		g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
			Chooser:  me.ID,
			Question: "Ureni — you may put a Dragon creature card from among them onto the battlefield",
			Cards:    ids,
			Min:      0,
			Max:      1,
			Zone:     game.ZoneLibrary,
		})
	})

	own := FilterViewFor(ViewOfGame(g), me.ID.String()).PendingChoices[0]
	if len(own.Options) != 3 {
		t.Fatalf("chooser sees %d candidates, want 3", len(own.Options))
	}
	for _, o := range own.Options {
		if o.Name == "" || !o.KnownByYou {
			t.Error("the looker cannot read a card they looked at")
		}
	}

	theirs := FilterViewFor(ViewOfGame(g), them.ID.String())
	other := theirs.PendingChoices[0]
	if len(other.Options) != 0 || other.ChooseMax != 0 {
		t.Errorf("opponent sees %d candidates bounded %d..%d", len(other.Options), other.ChooseMin, other.ChooseMax)
	}
	top := me.Library.Cards[len(me.Library.Cards)-1]
	if top.IsKnownTo(them.ID) {
		t.Error("a look made the opponent a knower")
	}
	for _, seat := range theirs.Seats {
		for _, c := range seat.Library.Cards {
			if c.InstanceID == top.InstanceID.String() && strings.TrimSpace(c.Name) != "" {
				t.Errorf("the opponent's view of %s's library shows the looked-at card %q", seat.Name, c.Name)
			}
		}
	}
}
