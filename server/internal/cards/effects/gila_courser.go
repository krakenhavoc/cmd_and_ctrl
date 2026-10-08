package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Gila Courser — Creature — Lizard Mount {2}{R}:
//
//	"Whenever this creature attacks while saddled, exile the top card of
//	 your library. Until the end of your next turn, you may play that
//	 card.
//	 Saddle 1"
//
// The exile and its permission are Prosper's Mystic Arcanum's
// (b20ExileTopUntilEndOfNextTurn): "play", so a land is not stranded,
// and the duration is ADR 0063's "until the end of your next turn".
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "bc983c67-eb5d-4a85-a12e-a410f85e949e",
		Name:         "Gila Courser",
		Completeness: CompletenessFull,
		Activated:    []ActivatedAbility{Saddle(1)},
		Triggered: []game.TriggeredAbility{
			AttacksWhileSaddled("Gila Courser — exile the top card of your library; you may play it until the end of your next turn", func(g *game.Game, item *game.StackItem) error {
				return b20ExileTopUntilEndOfNextTurn(g, item, 1)
			}),
		},
	})
}
