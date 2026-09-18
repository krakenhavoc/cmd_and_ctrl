package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Maelstrom Colossus — Artifact Creature — Golem {8}, 7/7 (EDHREC
// rank 4502):
//
//	"Cascade (When you cast this spell, exile cards from the top of
//	 your library until you exile a nonland card that costs less. You
//	 may cast it without paying its mana cost. Put the exiled cards on
//	 the bottom in a random order.)"
//
// Eight generic mana for a 7/7 and a free spell of mana value seven
// or less — which in a big-mana deck is very nearly "cast your best
// card". The COLOURLESS cost is why it is in this many lists: every
// deck can play it, so the highest-variance cascade in the format is
// available to all of them.
//
// Cascade is the keyword constructor. The engine owns the whole
// mechanic — the exile loop, the "you may cast it" prompt, the free
// cast and the random bottoming — so the Spec is one line.
//
// Note what cascade triggers on: CASTING, not resolving. A countered
// Colossus has already cascaded, and the cascaded spell resolves
// first because it goes on the stack above. FromStack on the
// constructor is what makes the trigger reachable at all — the source
// is on the stack, minutes of game time before it would reach the
// battlefield where the harvester scans.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "06fc3eb1-7a48-42a0-8a5c-d61bd5aacfff",
		Name:         "Maelstrom Colossus",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{Cascade()},
	})
}
