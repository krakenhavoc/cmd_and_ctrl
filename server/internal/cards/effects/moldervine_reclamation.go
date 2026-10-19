package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Moldervine Reclamation — Enchantment, {3}{B}{G} (EDHREC rank 981):
//
//	"Whenever a creature you control dies, you gain 1 life and draw a
//	 card."
//
// The aristocrats deck's card-draw engine. One dies trigger, gated
// to creatures the controller controlled — read off the dead card
// post-move, where the controller survives the trip to the graveyard
// — and paying out a life and a card in the printed order. An
// opponent's creature and a bounce or exile are both silent.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "68639a3d-2192-4921-8298-c76bb0cd6b02",
		Name:         "Moldervine Reclamation",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverACreatureYouControlDies("Moldervine Reclamation — gain 1 life and draw a card", Do(
				// The draw waits for a prompt the life gain asks (CR 608.2c, #4155).
				GainLife{Amount: 1},
				DrawCards{N: 1},
			)),
		},
	})
}
