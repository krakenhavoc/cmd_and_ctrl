package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bootleggers' Stash — Artifact {5}{G}:
//
//	"Lands you control have "{T}: Create a Treasure token.""
//
// Chromatic Lantern's layer 6 grant (ADR 0093), with an ordinary activated
// ability in the bundle instead of a mana ability: creating a Treasure is
// not a mana ability (CR 605.1a), so it uses the stack, and each land
// taps as its own "this" (CR 302.6 reads the land, not the Stash). The
// grant is to lands you control as they are now, read off the effective
// types.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d1deba03-b26a-4a0d-b242-24bb3a3f9ea6",
		Name:         "Bootleggers' Stash",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key: bootleggersStashTreasure,
			Activated: []effects.ActivatedAbility{{
				Label:   "{T}: Create a Treasure token.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    effects.TapCost(),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.CreateToken{Controller: item.Controller, Template: effects.TreasureToken(), N: 1}.Apply(effects.NewContext(g, item))
				},
			}},
			Text: "{T}: Create a Treasure token.",
		}},
		Static: []game.StaticAbility{
			effects.Labelled(effects.GrantAbilities(effects.LandsYouControl, bootleggersStashTreasure), "Lands you control have \"{T}: Create a Treasure token.\""),
		},
	})
}

// bootleggersStashTreasure is the granted bundle's key. An on-disk key.
const bootleggersStashTreasure = "bootleggers-stash/treasure"
