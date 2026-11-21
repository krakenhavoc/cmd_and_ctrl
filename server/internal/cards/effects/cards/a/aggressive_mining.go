package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Aggressive Mining — Enchantment {3}{R}:
//
//	"You can't play lands.
//	 Sacrifice a land: Draw two cards. Activate only once each turn."
//
// ADR 0109 §4 (#1895). "You can't play lands" is a LandPlayRestriction
// read off the permanent's CONTROLLER, so an Aggressive Mining an
// opponent has stolen restricts them. The ability is a sacrifice-a-land
// cost (Zuran Orb's) with CR 602.1b's once-a-turn limit read off the
// activation tally, so the second activation is never offered rather
// than paid for and then refused.
//
// No simplification.
func init() {
	const label = "Sacrifice a land: Draw two cards. Activate only once each turn."
	effects.Register(effects.Spec{
		OracleID:     "ff321c13-03ae-4cb9-b371-1242290f8433",
		Name:         "Aggressive Mining",
		Completeness: effects.CompletenessFull,
		LandPlayRestrictions: []game.LandPlayRestriction{
			effects.YouCantPlayLands("You can't play lands."),
		},
		Activated: []effects.ActivatedAbility{{
			Label:     label,
			Cost:      effects.B08SacrificeALand(),
			Condition: effects.OncePerTurnActivation(label),
			Purpose:   game.Purpose{Answers: game.AnswerValue},
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.DrawCards{Player: item.Controller, N: 2}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
