package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sedris, the Traitor King — Legendary Creature — Zombie Warrior
// {3}{U}{B}{R}:
//
//	"Each creature card in your graveyard has unearth {2}{B}. ({2}{B}:
//	 Return the card to the battlefield. The creature gains haste. Exile
//	 it at the beginning of the next end step or if it would leave the
//	 battlefield. Unearth only as a sorcery.)"
//
// A graveyard grant (#3233, ADR 0093's amendment of 2026-10-13) of the
// ordinary Unearth row. The recipient is a card in the controller's
// graveyard (CR 108.4, an owner's), and the row goes the pass after
// Sedris leaves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "48054407-b5e1-431c-ad03-c12a31d0d9a8",
		Name:         "Sedris, the Traitor King",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key:       "sedris-the-traitor-king/unearth",
			Zone:      game.ZoneGraveyard,
			Activated: []effects.ActivatedAbility{effects.Unearth("{2}{B}")},
			Text:      "Unearth {2}{B}",
		}},
		Static: []game.StaticAbility{
			effects.Labelled(effects.GrantAbilitiesInZone(game.ZoneGraveyard, effects.CardsYouOwn(effects.CreatureCard), "sedris-the-traitor-king/unearth"),
				"Each creature card in your graveyard has unearth {2}{B}."),
		},
	})
}
