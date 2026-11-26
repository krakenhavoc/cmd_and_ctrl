package cardssh

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Siren of the Silent Song — {1}{U}{B} Creature — Zombie Siren 2/1 (Born of the Gods):
//
//	"Flying
//	 Inspired — Whenever this creature becomes untapped, each opponent discards a card,
//	 then each opponent mills a card."
//
// Sheoldred's chapter II shape: every opponent answers their own discard first, and the
// mills happen once all of them have, in turn order.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "abb9a164-34cd-41d7-8e89-5700fff336dd",
		Name:            "Siren of the Silent Song",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventUntapCard, effects.Self, "Siren of the Silent Song — each opponent discards a card, then each opponent mills a card (inspired)",
				func(g *game.Game, item *game.StackItem) error {
					return g.EachPlayerDiscardsThenKeyedForEffect(item.Controller,
						game.DiscardPrompt{Source: item.SourceCardID, N: 1},
						sirenOfTheSilentSongMill, []uuid.UUID{item.Controller})
				}),
		},
	})
}

// sirenOfTheSilentSongMill is the mill, once every opponent has discarded. An on-disk
// key: never renamed, never reused.
var sirenOfTheSilentSongMill = game.RegisterPromptRunThen("prompt-run/siren-of-the-silent-song-mill", func(g *game.Game, r game.RunDone) error {
	if len(r.Carry) == 0 {
		return nil
	}
	c := effects.KeyedContext(g, r.Carry[0], r.Source)
	return effects.MillEachPlayerInTurn(c, c.Opponents(), effects.ConstantMill(1), game.KeyedTail{})
})
