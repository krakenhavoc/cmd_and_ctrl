package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Eager Construct — Artifact Creature — Construct {2}, 2/2 (Foundations
// Jumpstart):
//
//	"When this creature enters, each player may scry 1."
//
// Each living player in turn order from the active player is asked "scry 1?"
// and a yes scries before the next is asked (Model of Unity's ask-in-turn,
// eagerConstructAsk).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "39363190-7f79-4fae-bcfc-b17ec4e07bfe",
		Name:         "Eager Construct",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Scry: 1},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Eager Construct — each player may scry 1", func(g *game.Game, item *game.StackItem) error {
				return eagerConstructAsk(NewContext(g, item), apnapLivingSeats(g))
			}),
		},
	})
}
