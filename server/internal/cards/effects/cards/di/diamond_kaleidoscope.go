package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Diamond Kaleidoscope — Artifact {4} (Visions, S82 OLD, #5631):
//
//	"{3}, {T}: Create a 0/1 colorless Prism artifact creature token.
//	 Sacrifice a Prism token: Add one mana of any color."
//
// The sacrifice is a mana ability (CR 605.1a) like Phyrexian Altar's, over
// Prism tokens only. A Prism is no Treasure and the token's own subtype is
// what the cost names.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bffa9256-aff4-476f-86f4-4f4c9e57fa69",
		Name:         "Diamond Kaleidoscope",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}, {T}: Create a 0/1 colorless Prism artifact creature token.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Tokens: 1},
			Cost:    effects.Plus(effects.ManaCost("{3}"), effects.TapCost()),
			Effect:  effects.Do(effects.CreateToken{Template: effects.TokenCard("0/1 colorless Prism artifact"), N: 1}),
		}},
		ManaAbilities: []effects.ManaAbility{{
			Cost: effects.ManaAbilityCost{
				SacrificeOther: effects.SacrificeSpec("a Prism token", effects.HasSubtype("Prism"), effects.S82OldB788IsToken),
			},
			Produced: "{W|U|B|R|G}",
			Label:    "Sacrifice a Prism token: Add one mana of any color",
			Answers:  game.AnswerSacOutlet,
		}},
	})
}
