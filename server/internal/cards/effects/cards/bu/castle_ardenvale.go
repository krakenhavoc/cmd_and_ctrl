package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Castle Ardenvale — Land (EDHREC rank 708):
//
//	"This land enters tapped unless you control a Plains.
//	 {T}: Add {W}.
//	 {2}{W}{W}, {T}: Create a 1/1 white Human creature token."
//
// A white source that makes bodies when the hand is empty. The
// enters-tapped clause is the checkland shape with one land type
// instead of two (any land with the Plains subtype counts — a Plains,
// a Godless Shrine, a Savannah); the token ability is a CR 602
// activated ability with a mana and a tap component, making the same
// 1/1 Human Stroke of Midnight hands out.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f8f4fc60-725d-46d8-8e8f-e68e00d20589",
		Name:         "Castle Ardenvale",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.B06EntersTappedUnlessLandType("plains")},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{W}",
			Label:    "Add {W}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{W}{W}, {T}: Create a 1/1 white Human creature token.",
			Purpose: game.Purpose{Answers: game.AnswerMakesBlocker},
			Cost:    effects.Plus(effects.ManaCost("{2}{W}{W}"), effects.TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.CreateToken{
					Controller: item.Controller,
					Template:   effects.TokenCard("1/1 white Human"),
					N:          1,
				}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
