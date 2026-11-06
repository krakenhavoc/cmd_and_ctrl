package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mechan Navigator — Artifact Creature — Robot Pilot {1}{U}, 2/1 (Edge of
// Eternities, #4461):
//
//	"Whenever this creature becomes tapped, draw a card, then discard a
//	 card."
//
// Chrome Companion's "becomes tapped" watch (ThisBecameTapped over a tap
// and an attack, so a vigilance attacker does not loot) with Bellowing
// Crier's loot (b35LootOne): the draw first, then the discard, so the
// card just drawn can be the one discarded. Tapping it for a cost (crew,
// station, a convoke-style payment) loots too.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "fa72c607-7eb4-41b7-9af1-f3d6c69abf09",
		Name:         "Mechan Navigator",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			OnAny([]game.EventKind{game.EventTapCard, game.EventAttack}, ThisBecameTapped,
				"Mechan Navigator — draw a card, then discard a card", b35LootOne),
		},
	})
}
