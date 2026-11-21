package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// reveal_frame_test.go — CR 701.20 reveal, from the card end.
//
// The engine and protocol halves are pinned in their own packages.
// What is tested here is the thing those two cannot see: that the
// cards which PRINT the word "reveal" actually reach the broadcast
// frame, and that they reach it with the right cards named.
//
// Three shapes, one per way a catalog card reveals today:
//
//   - a tutor's "search your library for a card, REVEAL it, put it
//     into your hand" — the largest family by far, 59 card files;
//   - "reveal cards from the top until X" where the run ends in the
//     graveyard (Consuming Aberration);
//   - the same, where one card peels off to hand (Hermit Druid).

// revealsFor returns the reveal window as the named seat receives it,
// through the same per-viewer filter a real client is served by.
func RevealsFor(g *game.Game, viewer uuid.UUID) []protocol.RevealView {
	return protocol.FilterViewFor(protocol.ViewOfGame(g), viewer.String()).Reveals
}

// revealedNames flattens a reveal window into the card names it
// announced, oldest reveal first.
func RevealedNames(rs []protocol.RevealView) []string {
	var out []string
	for _, r := range rs {
		for _, c := range r.Cards {
			out = append(out, c.Name)
		}
	}
	return out
}

func ContainsName(names []string, want string) bool {
	for _, n := range names {
		if n == want {
			return true
		}
	}
	return false
}

// castVanillaSorceryForRevealTest puts a plain sorcery into a
// player's hand and casts it, purely to fire a "whenever you cast a
// spell" trigger.
func CastVanillaSorceryForRevealTest(t *testing.T, g *game.Game, p *game.Player) {
	t.Helper()
	id := uuid.New()
	g.WithWriteLock(func() {
		p.Hand.PushTop(game.Card{
			InstanceID: id,
			Name:       "Trigger Bait",
			TypeLine:   "Sorcery",
			Owner:      p.ID,
			Controller: p.ID,
		})
	})
	if err := g.CastSpell(p.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
}
