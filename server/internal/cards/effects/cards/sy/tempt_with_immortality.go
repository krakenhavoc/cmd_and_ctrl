package cardssy

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tempt with Immortality — Sorcery {4}{B} (Secret Lair Drop):
//
//	"Tempting offer — Return a creature card from your graveyard to the
//	 battlefield. Each opponent may return a creature card from their
//	 graveyard to the battlefield. For each opponent who does, return a
//	 creature card from your graveyard to the battlefield."
//
// The shared tempting offer (NewCmdr2S49Tempting): the caster returns one
// first, each opponent in turn order is asked, and an opponent who accepts
// returns one of their own and the caster returns another. Each return is a
// chosen card (not a target) from the returning player's own graveyard.
func init() {
	tempt := effects.NewCmdr2S49Tempting("confirm/tempt-with-immortality",
		"Tempt with Immortality — return a creature card from your graveyard to the battlefield (the caster returns another if you do)?",
		func(ctx *effects.Context, player uuid.UUID) error {
			return effects.ReturnChosenFromGraveyard{
				Player:   player,
				Question: "Tempt with Immortality — return a creature card from your graveyard to the battlefield",
				Match:    game.Card.IsCreature,
				Min:      1, Max: 1,
			}.Apply(ctx)
		})
	effects.Register(effects.Spec{
		OracleID:     "06e1c0fa-767c-4204-972f-d98f770d85f3",
		Name:         "Tempt with Immortality",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerGraveyard},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return tempt.Resolve(ctx)
		},
	})
}
