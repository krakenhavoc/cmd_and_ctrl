package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Veteran Soldier — Legendary Enchantment — Background {1}{W} (Commander
// Legends: Battle for Baldur's Gate, #5948):
//
//	"Commander creatures you own have "Whenever this creature attacks a player,
//	 if no opponent has more life than that player, for each opponent, create a
//	 1/1 white Soldier creature token that's tapped and attacking that
//	 opponent.""
//
// Guild Artisan's granted trigger (an attack on a player; an intervening "if"
// read as the attack is declared and again as the trigger resolves, CR 603.4)
// with Adeline's per-opponent tapped-and-attacking tokens: one Soldier for each
// opponent still in the game as it resolves, attacking that player, whether or
// not that player was attacked. The tokens were never declared as attackers, so
// they fire nothing that watches attacks (CR 508.4). Choose a Background is
// deck construction (ADR 0144).
//
// No simplification.
const veteranSoldierGrant = "veteran-soldier/soldiers"

func init() {
	effects.Register(effects.Spec{
		OracleID:     "6bbf1410-800d-4d05-93d2-ba3c1656728e",
		Name:         "Veteran Soldier",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key: veteranSoldierGrant,
			Triggered: []game.TriggeredAbility{
				effects.WheneverThisAttacksAPlayerNoOpponentRicher("Veteran Soldier — for each opponent, create a 1/1 white Soldier creature token that's tapped and attacking that opponent",
					func(g *game.Game, item *game.StackItem) error {
						return effects.TokenTappedAndAttackingEachOpponent(effects.NewContext(g, item), func() game.Card {
							return effects.TokenCard("1/1 white Soldier")
						})
					}),
			},
			Text: "Whenever this creature attacks a player, if no opponent has more life than that player, for each opponent, create a 1/1 white Soldier creature token that's tapped and attacking that opponent.",
		}},
		Static: []game.StaticAbility{effects.GrantToCommanderCreaturesYouOwn(veteranSoldierGrant)},
	})
}
