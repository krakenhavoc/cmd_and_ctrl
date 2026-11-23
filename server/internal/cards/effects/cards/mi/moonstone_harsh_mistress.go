package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Moonstone, Harsh Mistress — Legendary Creature — Human Doctor Villain
// {3}{B}, 2/4 (Marvel Super Heroes, #4455):
//
//	"Flying
//	 Whenever you discard a card, you may exile that card from your
//	 graveyard. If you do, until the end of your next turn, you may play
//	 that card."
//
// Containment Construct's trigger (one per discarded card, optional, the
// card exiled only if it is still in the graveyard) with the grant lasting
// until the end of the controller's next turn (ADR 0063's seat-turn
// duration) instead of this turn. The grant is "play", so a discarded land
// can be played off it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "aeeb50f2-8e02-4d40-a1dd-28b1c5c85c7e",
		Name:            "Moonstone, Harsh Mistress",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventDiscardCard},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return effects.DiscardedByYou(ev, source)
			},
			OptionalPrompt: &game.TriggerOptionalPrompt{Question: "Moonstone, Harsh Mistress — exile the discarded card to play it until the end of your next turn?"},
			Key:            "Moonstone, Harsh Mistress — exile the discarded card, play it until the end of your next turn",
			Effect: func(g *game.Game, item *game.StackItem) error {
				id := item.Trigger.Event.CardID
				if !effects.DiscardedCardStillInGraveyard(g, item, id) {
					return nil
				}
				return g.ExileCardWithPermissionForEffect(id, game.CastPermission{
					Player:   item.Controller,
					Duration: g.UntilEndOfYourNextTurnDuration(item.Controller),
				})
			},
		}},
	})
}
