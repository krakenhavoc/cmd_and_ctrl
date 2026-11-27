package cardsmi

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mossdog — Creature — Plant Dog {G}, 1/1 (Nemesis, S82 OLD batch 680,
// #5618):
//
//	"Whenever this creature becomes the target of a spell or ability an
//	 opponent controls, put a +1/+1 counter on this creature."
//
// The becomes-target event (one per target slot, CR 115.3) names the player
// who controls the spell or ability: an opponent's triggers it, the
// controller's own does not.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "95cc3652-b78b-473b-84cf-0dd3a7006219",
		Name:         "Mossdog",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBecomesTarget, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Kind == game.EventBecomesTarget && ev.CardID == source.InstanceID && ev.Actor != source.Controller && ev.Actor != uuid.Nil
			}, "Mossdog — put a +1/+1 counter on this creature",
				effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
