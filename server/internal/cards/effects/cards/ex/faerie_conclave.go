package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Faerie Conclave — Land:
//
//	"This land enters tapped.
//	 {T}: Add {U}.
//	 {1}{U}: This land becomes a 2/1 blue Faerie creature with flying
//	 until end of turn. It's still a land."
//
// A manland built like Creeping Tar Pit, through cmdr2B13Animate.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0c25f6b1-8fb3-4406-9605-0282d2dbbcec",
		Name:         "Faerie Conclave",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SelfEntersTapped()},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{U}",
			Label:    "Add {U}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{U}: This land becomes a 2/1 blue Faerie creature with flying until end of turn. It's still a land.",
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Cost:    effects.ManaCost("{1}{U}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !effects.SourceIsStillThisPermanent(g, item) {
					return nil
				}
				return effects.Cmdr2B13Animate(effects.NewContext(g, item), item.SourceCardID,
					"Faerie Conclave — a 2/1 blue Faerie creature with flying until end of turn",
					2, 1, "Faerie", []string{"U"}, "flying")
			},
		}},
	})
}
