package heuristic_test

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// #2591. Manifest dread's prompt is a choose_cards over two looked-at
// library cards, so the generic library-take rule took the BEST card
// and manifested it, even a noncreature, which can never be turned face
// up (CR 701.40b), while the creature went to the graveyard.
func TestManifestDreadManifestsACreature(t *testing.T) {
	const choiceID = "choice-dread"
	const reason = "Manifest dread — put one onto the battlefield face down; the other goes to your graveyard"
	pick := func(t *testing.T, a, b protocol.CardView) string {
		t.Helper()
		v := newView([]protocol.PlayerView{newSeat(0, withLibrary(a, b)), newSeat(1)},
			withBattlefield(sixLands()...),
			withChoice(libraryChoice(choiceID, reason, 1, 1, a, b)))
		in := input(0, v,
			choiceMove(t, 0, choiceID, "manifest "+a.Name, ids(a.InstanceID)),
			choiceMove(t, 0, choiceID, "manifest "+b.Name, ids(b.InstanceID)),
		)
		return chose(t, in, decide(t, heuristic.New(), in))
	}
	artifact := func(id, name, cost string) protocol.CardView {
		c := spell(id, 0, name, cost)
		c.TypeLine = "Artifact"
		return c
	}

	t.Run("a creature beats a pricier noncreature, in either order", func(t *testing.T) {
		bear := creature(cardID(1), 0, "Bear", 2, 2)
		rock := artifact(cardID(2), "Big Rock", "{6}")
		if got := pick(t, bear, rock); got != "manifest Bear" {
			t.Fatalf("chose %q, want the creature", got)
		}
		if got := pick(t, rock, bear); got != "manifest Bear" {
			t.Fatalf("chose %q, want the creature", got)
		}
	})

	t.Run("two creatures: the cheaper to turn face up", func(t *testing.T) {
		cheap := creature(cardID(1), 0, "Bear", 2, 2)
		dear := creature(cardID(2), 0, "Dragon", 6, 6)
		dear.ManaCost = "{4}{R}{R}"
		if got := pick(t, dear, cheap); got != "manifest Bear" {
			t.Fatalf("chose %q, want the cheaper creature", got)
		}
	})

	t.Run("no creature: manifest the lesser card", func(t *testing.T) {
		small := artifact(cardID(1), "Trinket", "{1}")
		big := artifact(cardID(2), "Big Rock", "{6}")
		if got := pick(t, big, small); got != "manifest Trinket" {
			t.Fatalf("chose %q, want the lesser card", got)
		}
	})
}
