package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Containment Construct — Artifact Creature — Construct {2}, 2/1
// (EDHREC rank 1854):
//
//	"Whenever you discard a card, you may exile that card from your
//	 graveyard. If you do, you may play that card this turn."
//
// The discard deck's value engine: every looted, rummaged or
// wheeled card is playable this turn instead of gone. One trigger
// per discarded card (the engine emits EventDiscardCard per card),
// optional, so the controller is asked before it goes on the stack.
// On resolution the card is exiled from the graveyard — only if it
// is still there (a Containment Construct trigger responded to with
// a reanimation does nothing) — carrying a "play" grant until end
// of turn, so a discarded LAND can be played off it. Timing still
// applies: the grant says you may, not when.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "54d48e54-be4a-4a78-a778-b28e74ef7134",
		Name:         "Containment Construct",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventDiscardCard},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return discardedByYou(ev, source)
			},
			OptionalPrompt: &game.TriggerOptionalPrompt{Question: "Containment Construct — exile the discarded card to play it this turn?"},
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				discarded := ev.CardID
				return game.NewTriggeredItem(source, "Containment Construct — exile the discarded card, play it this turn",
					func(g *game.Game, item *game.StackItem) error {
						return b17ExileFromGraveyardAndMayPlay(g, item.Controller, discarded)
					})
			},
		}},
	})
}
