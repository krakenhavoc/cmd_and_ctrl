package cardsbl

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Brackish Trudge — Creature — Fungus Beast {2}{B}, 4/2 (Strixhaven: School of
// Mages, #5502):
//
//	"This creature enters tapped.
//	 {1}{B}: Return this card from your graveyard to your hand. Activate only
//	 if you gained life this turn."
//
// The graveyard ability works only while the card is in the graveyard (CR
// 113.6) and only on a turn its controller gained life, read from the turn's
// life-gain tally as the activation is announced (CR 602.5); it does nothing
// if the card has left by resolution (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c775d3ef-2184-4425-8c8e-2bbc49687f76",
		Name:         "Brackish Trudge",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SelfEntersTapped()},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{B}: Return this card from your graveyard to your hand. Activate only if you gained life this turn.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.ManaCost("{1}{B}"),
			Zones:   []game.ZoneKind{game.ZoneGraveyard},
			Condition: func(g *game.Game, controller, _ uuid.UUID) bool {
				return g.TurnTallyFor(controller).LifeGained > 0
			},
			Effect: effects.ReturnThisFromYourGraveyardToHand,
		}},
	})
}
