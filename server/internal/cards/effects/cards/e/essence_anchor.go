package cardse

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Essence Anchor — Artifact {2}{U} (Tarkir: Dragonstorm, #4463):
//
//	"At the beginning of your upkeep, surveil 1.
//	 {T}: Create a 2/2 black Zombie Druid creature token. Activate only
//	 during your turn and only if a card left your graveyard this turn."
//
// The upkeep trigger is the footprinted Surveil primitive. The activation
// condition is two clauses of one sentence, both read as the ability is
// activated: it is the controller's turn, and a card they own has left a
// graveyard this turn (cardLeftYourGraveyardThisTurn, Gau's and Bonecache
// Overseer's reading of the turn's events: a cast from the graveyard and a
// card exiled from it both count). The token is the Zombie Druid of The
// Sibsig Ceremony.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e28a3f0e-40bd-4293-b5eb-ec9d185c1e1c",
		Name:         "Essence Anchor",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1},
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Essence Anchor — surveil 1", effects.Do(effects.Surveil{N: 1})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:     "{T}: Create a 2/2 black Zombie Druid creature token. Activate only during your turn and only if a card left your graveyard this turn.",
			Purpose:   game.Purpose{Tokens: 1, Answers: game.AnswerValue},
			Cost:      effects.TapCost(),
			Condition: essenceAnchorCanActivate,
			Effect:    effects.Do(effects.CreateToken{Template: effects.TokenCard("2/2 black Zombie Druid"), N: 1}),
		}},
	})
}

// essenceAnchorCanActivate is "only during your turn and only if a card
// left your graveyard this turn".
func essenceAnchorCanActivate(g *game.Game, controller, _ uuid.UUID) bool {
	return effects.IsYourTurn(g, controller) && effects.CardLeftYourGraveyardThisTurn(g, controller)
}
