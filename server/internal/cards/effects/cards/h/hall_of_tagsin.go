package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hall of Tagsin — Land (The Brothers' War, #5482):
//
//	"{T}: Add {C}.
//	 {1}, {T}: Add one mana of any color.
//	 {4}, {T}: Create a tapped Powerstone token."
//
// Capital City's two mana abilities, plus an activated ability (it
// makes a token, so it is not a mana ability, CR 605.1a) that taps the
// land and makes a tapped Powerstone.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fd809587-773c-47ec-8676-cefef6d1e38f",
		Name:         "Hall of Tagsin",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{
			effects.PainlessColorless(),
			{
				Cost:     effects.ManaAbilityCost{Tap: true, Mana: "{1}"},
				Produced: "{W|U|B|R|G}",
				Label:    "{1}, {T}: Add one mana of any color",
			},
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{4}, {T}: Create a tapped Powerstone token.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{4}"), effects.TapCost()),
			Effect:  effects.Do(effects.CreateTokenAdvanced{Spec: effects.Token(effects.PowerstoneToken()).EntersTapped(), N: 1}),
		}},
	})
}
