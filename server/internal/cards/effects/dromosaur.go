package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dromosaur — Creature — Dinosaur {2}{R}, 2/3 (Urza's Saga):
//
//	"Whenever this creature blocks or becomes blocked, it gets +2/-2
//	 until end of turn."
//
// Bushido's trigger with +2/-2: "blocks" triggers once however many
// attackers it blocks (CR 509.3a), and "becomes blocked" once however
// many creatures block it (CR 509.3c).
//
// No simplification.
func init() {
	const label = "Dromosaur — +2/-2 until end of turn"
	Register(Spec{
		OracleID:     "4576cc5b-7b3a-41a7-9b84-acadf2d4adf5",
		Name:         "Dromosaur",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			OnAny([]game.EventKind{game.EventBlock, game.EventBecomesBlocked}, thisBlocksOrBecomesBlocked,
				label, thisGetsUntilEndOfTurn(2, -2, label)),
		},
	})
}
