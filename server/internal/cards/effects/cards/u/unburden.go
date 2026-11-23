package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Unburden — Sorcery {1}{B}{B} (Scourge):
//
//	"Target player discards two cards.
//	 Cycling {2} ({2}, Discard this card: Draw a card.)"
//
// Mind Rot's discard (the player chooses), and cycling as the keyword's own
// hand ability (CR 702.29a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "97acd3aa-ae4a-4e9d-a5b9-36eaba158735",
		Name:         "Unburden",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target player"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			player, ok := effects.FirstLegalPlayerTarget(ctx)
			if !ok {
				return nil
			}
			ctx.Game.QueueDiscardChoiceForEffect(game.DiscardPrompt{
				Player: player,
				Source: item.SourceCardID,
				N:      2,
			})
			return nil
		},
		Activated: []effects.ActivatedAbility{effects.Cycling("{2}")},
	})
}
