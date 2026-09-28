package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Goblin Grappler — Creature — Goblin, {R}, 1/1:
//
//	"Provoke (Whenever this creature attacks, you may have target
//	 creature defending player controls untap and block it if able.)"
//
// #1684: the plainest Provoke (CR 702.39) in print — the keyword and
// nothing else. Provoke() untaps the target and registers a
// blocksAttacker requirement naming the Grappler as the object that
// attacked.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "48c1c84c-f690-42ce-9c5a-dd09b1e4197c",
		Name:         "Goblin Grappler",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{Provoke()},
	})
}
