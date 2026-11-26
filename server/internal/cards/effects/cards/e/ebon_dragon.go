package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ebon Dragon — Creature — Dragon {5}{B}{B}, 5/4 (From the Vault: Dragons, S82
// OLD batch 519, #5589):
//
//	"Flying
//	 When this creature enters, you may have target opponent discard a card."
//
// Nath of the Gilt-Leaf's shape: the opponent is the trigger's target, chosen
// as it goes on the stack, and the "may" is answered as it resolves (CR 603.5).
// The discard is that opponent's own choice of card (CR 701.9a, not random), a
// hand with nothing in it discarding nothing, and a target gone by resolution
// makes the ability do nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "12e4d2bd-83e9-4120-a2e8-0645c0ed2387",
		Name:            "Ebon Dragon",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.Targeting(effects.WhenThisEnters("Ebon Dragon — you may have target opponent discard a card",
				func(g *game.Game, item *game.StackItem) error {
					player := effects.TargetedPlayer(effects.NewContext(g, item))
					g.QueueDiscardChoiceForEffect(game.DiscardPrompt{Player: player, Source: item.SourceCardID, N: 1})
					return nil
				}),
				effects.TargetPlayer("target opponent", effects.Opponent())),
				"Ebon Dragon — have target opponent discard a card?"),
		},
	})
}
