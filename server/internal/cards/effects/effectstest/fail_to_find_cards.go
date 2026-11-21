package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// fail_to_find_cards_test.go — #3053 / #3081 through real catalog
// cards. The engine rule (CR 701.23b/d) is pinned in the game package;
// this file pins that the cards inherit it without a per-card flag:
//
//   - a basic-land fetch against exactly one basic still asks, and
//     "find nothing" is honoured with the shuffle intact;
//   - the "up to N" cards of #3081 may take fewer than N (and 0);
//   - Demonic Tutor and Intuition — no stated quality — must find.

const DemonicTutorOracle = "82004860-e589-4e38-8d61-8c0210e4ea39"

func SearchEventCount(g *game.Game) int {
	n := 0
	g.ReadSnapshot(func() {
		for _, ev := range g.Events {
			if ev.Kind == game.EventSearchLibrary {
				n++
			}
		}
	})
	return n
}

// libraryKnownTo counts library cards the seat still knows the identity
// of: zero after a shuffle (the shuffle un-knows the zone).
func LibraryKnownTo(p *game.Player, who uuid.UUID) int {
	n := 0
	for _, c := range p.Library.Cards {
		if c.KnownBy[who] {
			n++
		}
	}
	return n
}
