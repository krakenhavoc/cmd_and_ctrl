package cardsap

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ashen Ghoul — Creature — Zombie {3}{B}, 3/1 (Ice Age, S82 OLD batch 851,
// #5636, seam #6194):
//
//	"Haste
//	 {B}: Return this card from your graveyard to the battlefield. Activate
//	 only during your upkeep and only if three or more creature cards are
//	 above this card."
//
// Bloodsoaked Champion's activated ability of the card in the graveyard, gated
// on your upkeep and on the graveyard's order (#6194,
// Game.CountGraveyardCardsAboveForEffect): the cards above this one in the pile,
// counting creature cards only. The gate is checked as the ability is
// activated (CR 602.5b); the return is read again as it resolves that the card
// is still in the graveyard.
//
// No simplification.
func init() {
	duringYourUpkeep := effects.DuringYourUpkeep()
	effects.Register(effects.Spec{
		OracleID:        "46fbcbdf-73d0-4fa2-8db7-3b19bfc7ac22",
		Name:            "Ashen Ghoul",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{B}: Return this card from your graveyard to the battlefield. Activate only during your upkeep and only if three or more creature cards are above this card.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.ManaCost("{B}"),
			Zones:   []game.ZoneKind{game.ZoneGraveyard},
			Condition: func(g *game.Game, controller, source uuid.UUID) bool {
				return duringYourUpkeep(g, controller, source) &&
					g.CountGraveyardCardsAboveForEffect(source, game.Card.IsCreature) >= 3
			},
			Effect: effects.ReturnThisFromGraveyardToBattlefield,
		}},
	})
}
