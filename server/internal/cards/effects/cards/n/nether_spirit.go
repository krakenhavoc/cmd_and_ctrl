package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Nether Spirit — Creature — Spirit {1}{B}{B}, 2/2 (Modern Horizons):
//
//	"At the beginning of your upkeep, if this card is the only creature card
//	 in your graveyard, you may return this card to the battlefield."
//
// A graveyard trigger (CR 113.6k) with an intervening "if" (CR 603.4), read as
// the upkeep begins and again as the ability resolves: Nether Spirit must be
// the one creature card there. "You may" is asked after the condition holds.
// The return is ReturnThisCardFromYourGraveyard, which does nothing for a card
// that has left the graveyard since (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "194e2563-c0bf-4151-bbb0-0d626130b055",
		Name:         "Nether Spirit",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.InGraveyard(effects.OldB165IfThisIsTheOnlyCreatureCardInYourGraveyard(
				effects.AtYourUpkeep("Nether Spirit — return it to the battlefield", effects.ReturnThisCardFromYourGraveyard))),
				"Nether Spirit — return it to the battlefield?"),
		},
	})
}
