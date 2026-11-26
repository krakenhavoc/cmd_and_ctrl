package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Avaricious Dragon — Creature — Dragon {2}{R}{R}, 4/4 (Magic Origins):
//
//	"Flying
//	 At the beginning of your draw step, draw an additional card.
//	 At the beginning of your end step, discard your hand."
//
// Two beginning-of-step triggers, one each, both only on its controller's
// turn. The discard is the whole hand with no choice; an empty hand
// discards nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "9be9d534-470d-463a-ba2f-9817a6723692",
		Name:            "Avaricious Dragon",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBeginDrawStep, effects.ByYou,
				"Avaricious Dragon — draw an additional card", effects.Do(effects.DrawCards{N: 1})),
			effects.On(game.EventBeginEndStep, effects.ByYou,
				"Avaricious Dragon — discard your hand", func(g *game.Game, item *game.StackItem) error {
					_, err := effects.DiscardWholeHand(g, item.Controller)
					return err
				}),
		},
	})
}
