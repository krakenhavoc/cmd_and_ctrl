package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Thunderscape Master — Creature — Human Wizard {2}{R}{R}, 2/2 (Invasion,
// #5616):
//
//	"{B}{B}, {T}: Target player loses 2 life and you gain 2 life.
//	 {G}{G}, {T}: Creatures you control get +2/+2 until end of turn."
//
// The drain is Syphon Life's; the pump is Righteous Charge's, applied to the
// creatures you control as the ability resolves (CR 611.2c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e1db1798-17d1-4ef3-80d7-3f9115a01f7b",
		Name:         "Thunderscape Master",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			effects.S82OldB659DrainTwoGainTwo(),
			{
				Label:   "{G}{G}, {T}: Creatures you control get +2/+2 until end of turn.",
				Purpose: game.Purpose{TeamPump: &game.Pump{Power: 2, Toughness: 2}, Answers: game.AnswerPump},
				Cost:    effects.Plus(effects.ManaCost("{G}{G}"), effects.TapCost()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.CreaturesYouControlGetUntilEOT(2, 2, "Thunderscape Master — creatures you control get +2/+2")(item, effects.NewContext(g, item))
				},
			},
		},
	})
}
