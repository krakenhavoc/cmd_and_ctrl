package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Needle Spires — Land:
//
//	"This land enters tapped.
//	 {T}: Add {R} or {W}.
//	 {2}{R}{W}: Until end of turn, this land becomes a 2/1 red and white
//	 Elemental creature with double strike. It's still a land."
//
// A manland built like Creeping Tar Pit, through cmdr2B13Animate.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e7bb8160-0a4b-4e46-b196-7a19fb388d8e",
		Name:         "Needle Spires",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SelfEntersTapped()},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{R|W}",
			Label:    "Add {R} or {W}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{R}{W}: Until end of turn, this land becomes a 2/1 red and white Elemental creature with double strike. It's still a land.",
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Cost:    effects.ManaCost("{2}{R}{W}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !effects.SourceIsStillThisPermanent(g, item) {
					return nil
				}
				return effects.Cmdr2B13Animate(effects.NewContext(g, item), item.SourceCardID,
					"Needle Spires — a 2/1 red and white Elemental creature with double strike until end of turn",
					2, 1, "Elemental", []string{"R", "W"}, "double strike")
			},
		}},
	})
}
