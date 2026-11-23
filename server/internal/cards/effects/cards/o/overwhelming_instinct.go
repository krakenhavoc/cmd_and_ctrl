package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Overwhelming Instinct — Enchantment {2}{G} (Starter Commander Decks):
//
//	"Whenever you attack with three or more creatures, draw a card."
//
// Armasaur Guide's and Meddling Youths' count: the EventAttack events of one
// declaration by you, settled as a batch, one trigger when there are at least
// three. An additional combat's declaration is its own batch.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9ba89c25-5302-414f-a2ec-436b8981999d",
		Name:         "Overwhelming Instinct",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WithAtBatchEnd(effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return source != nil && effects.AttackDeclaredByYou(ev, source.Controller)
			}, "Overwhelming Instinct — draw a card (you attacked with three or more creatures)",
				effects.Do(effects.DrawCards{N: 1})),
				effects.ThatManyEventsAtLeast(3)),
		},
	})
}
