package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cloakwood Hermit — Legendary Enchantment — Background {2}{G} (Commander
// Legends: Battle for Baldur's Gate, #5948):
//
//	"Commander creatures you own have "At the beginning of your end step, if
//	 a creature card was put into your graveyard from anywhere this turn,
//	 create two tapped 1/1 green Squirrel creature tokens.""
//
// A granted trigger (ADR 0093, Guild Artisan's recipient set), the commander
// creature's own: "your" is whoever controls it. The intervening "if" (CR
// 603.4) is read as the end step begins and again as the trigger resolves,
// off the turn's zone moves into a graveyard (a token is not a card; a card
// goes by its owner). Two tokens enter tapped. Choose a Background is deck
// construction (ADR 0144).
//
// No simplification.
const cloakwoodHermitGrant = "cloakwood-hermit/squirrels"

func init() {
	effects.Register(effects.Spec{
		OracleID:     "f898bf43-6a9c-46eb-b805-63cb12bab273",
		Name:         "Cloakwood Hermit",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key: cloakwoodHermitGrant,
			Triggered: []game.TriggeredAbility{
				effects.InterveningIfYou("a creature card was put into your graveyard from anywhere this turn",
					effects.CreatureCardWasPutIntoYourGraveyardThisTurn,
					effects.AtYourEndStep("Cloakwood Hermit — create two tapped 1/1 green Squirrel creature tokens",
						effects.Do(effects.CreateTokenAdvanced{Spec: effects.Token(effects.TokenCard("1/1 green Squirrel")).EntersTapped(), N: 2}))),
			},
			Text: "At the beginning of your end step, if a creature card was put into your graveyard from anywhere this turn, create two tapped 1/1 green Squirrel creature tokens.",
		}},
		Static: []game.StaticAbility{effects.GrantToCommanderCreaturesYouOwn(cloakwoodHermitGrant)},
	})
}
