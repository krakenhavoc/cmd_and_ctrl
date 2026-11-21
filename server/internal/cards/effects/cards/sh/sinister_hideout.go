package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sinister Hideout — Land (Marvel's Spider-Man):
//
//	"This land enters tapped.
//	 {T}: Add {U} or {B}.
//	 {4}, {T}: Surveil 1. (Look at the top card of your library. You may put
//	 it into your graveyard.)"
//
// Enters tapped is the self replacement (CR 614.1c); the mana ability is the
// declared two-colour one (the land has no basic land types to derive it
// from). The surveil is an ordinary activated ability: its cost is {4} and
// tapping the land, and a land that entered tapped can't pay it until it
// untaps.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "80f08b47-a237-4efd-8d86-dfe35a816b0e",
		Name:          "Sinister Hideout",
		Completeness:  effects.CompletenessFull,
		Replacements:  []game.ReplacementEffect{effects.SelfEntersTapped()},
		ManaAbilities: []effects.ManaAbility{effects.DualManaAbility("U", "B")},
		Activated: []effects.ActivatedAbility{{
			Label:   "{4}, {T}: Surveil 1.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{4}"), effects.TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.Surveil{Player: item.Controller, N: 1}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
