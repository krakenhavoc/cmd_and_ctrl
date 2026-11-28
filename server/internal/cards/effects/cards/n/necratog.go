package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Necratog — Creature — Atog {1}{B}{B}, 1/2 (Weatherlight, S82 older sets batch 778, #5629):
//
//	"Exile the top creature card of your graveyard: This creature gets +2/+2 until end of turn."
//
// The cost is #6194's top-of-graveyard exile (CR 404.2, game.ExileCost.TopOnly): the activator does not
// pick the card, it is the topmost creature card of their graveyard, and a graveyard with none can't pay
// (CR 118.3). Unlimited, with no mana.
//
// No simplification.
func init() {
	const label = "Exile the top creature card of your graveyard: This creature gets +2/+2 until end of turn."
	effects.Register(effects.Spec{
		OracleID:     "9f820f1e-3cd1-4a3e-8d00-194c8a80b296",
		Name:         "Necratog",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   label,
			Cost:    effects.ExileTopFromGraveyard(1, "the top creature card of your graveyard", effects.IsCreatureCard),
			Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Power: 2, Toughness: 2}},
			Effect:  effects.ThisGetsUntilEndOfTurn(2, 2, "Necratog — +2/+2"),
		}},
	})
}
