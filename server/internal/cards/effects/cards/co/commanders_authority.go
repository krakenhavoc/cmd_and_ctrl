package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Commander's Authority — {4}{W} Aura: "Enchant creature. Enchanted creature
// has "At the beginning of your upkeep, create a 1/1 white Human creature
// token.""
//
// "Your" is the enchanted creature's controller (a granted bundle).
//
// No simplification.
func init() {
	const grant = "commanders-authority/upkeep"
	effects.Register(effects.Spec{
		OracleID:     "57bc1106-81a0-4d91-94db-25e3a5f989cb",
		Name:         "Commander's Authority",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Grants: []effects.AbilityGrant{{
			Key: grant,
			Triggered: []game.TriggeredAbility{
				effects.AtYourUpkeep("Commander's Authority — create a 1/1 white Human creature token",
					effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 white Human"), N: 1})),
			},
			Text: "At the beginning of your upkeep, create a 1/1 white Human creature token.",
		}},
		Static: []game.StaticAbility{effects.GrantAbilitiesToAttached(grant)},
	})
}
