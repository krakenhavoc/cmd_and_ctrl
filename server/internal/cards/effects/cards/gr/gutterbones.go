package cardsgr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Gutterbones — Creature — Skeleton Warrior {B}, 2/1 (Ravnica Allegiance):
//
//	"This creature enters tapped.
//	 {1}{B}: Return this card from your graveyard to your hand. Activate only during
//	 your turn and only if an opponent lost life this turn."
//
// Brackish Trudge's shape: the ability is activated from the graveyard. Both
// restrictions are checked as it is activated; life lost counts from any source,
// this turn, by any opponent (CR 119.3).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "76ae3779-c821-4cd3-8d0f-8c42d1206ce4",
		Name:         "Gutterbones",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SelfEntersTapped()},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{B}: Return this card from your graveyard to your hand. Activate only during your turn and only if an opponent lost life this turn.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.ManaCost("{1}{B}"),
			Zones:   []game.ZoneKind{game.ZoneGraveyard},
			Condition: func(g *game.Game, controller, source uuid.UUID) bool {
				return effects.IsYourTurn(g, controller) && effects.OpponentLifeLostThisTurn(g, controller) > 0
			},
			Effect: effects.ReturnThisFromYourGraveyardToHand,
		}},
	})
}
