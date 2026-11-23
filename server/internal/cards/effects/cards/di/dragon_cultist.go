package cardsdi

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dragon Cultist — Legendary Enchantment — Background {4}{R} (Commander
// Legends: Battle for Baldur's Gate, #5948):
//
//	"Commander creatures you own have "At the beginning of your end step, if a
//	 source you controlled dealt 5 or more damage this turn, create a 4/4 red
//	 Dragon creature token with flying.""
//
// A granted trigger (ADR 0093), the commander creature's own, with an
// intervening "if" (CR 603.4) read as the end step begins and again as it
// resolves. "A source" is one source: its damage this turn, to players and
// permanents alike, is added up from the turn's damage events, attributed to
// the player who controlled the source as it dealt it, so a creature that has
// since died still counts. Choose a Background is deck construction (ADR 0144).
//
// No simplification.
const dragonCultistGrant = "dragon-cultist/dragon"

func init() {
	effects.Register(effects.Spec{
		OracleID:     "d1d7c5fa-a92a-4815-a558-6d3daf37f4c4",
		Name:         "Dragon Cultist",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key: dragonCultistGrant,
			Triggered: []game.TriggeredAbility{
				effects.InterveningIfYou("a source you controlled dealt 5 or more damage this turn",
					func(g *game.Game, you uuid.UUID) bool {
						return effects.ASourceYouControlledDealtAtLeastDamageThisTurn(g, you, 5)
					},
					effects.AtYourEndStep("Dragon Cultist — create a 4/4 red Dragon creature token with flying",
						effects.Do(effects.CreateToken{Template: effects.TokenCard("4/4 red Dragon with flying"), N: 1}))),
			},
			Text: "At the beginning of your end step, if a source you controlled dealt 5 or more damage this turn, create a 4/4 red Dragon creature token with flying.",
		}},
		Static: []game.StaticAbility{effects.GrantToCommanderCreaturesYouOwn(dragonCultistGrant)},
	})
}
