package cardssh

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Skitterskin — Creature — Eldrazi Drone {3}{B}, 4/3 (Battle for Zendikar,
// #5547):
//
//	"Devoid (This card has no color.)
//	 This creature can't block.
//	 {1}{B}: Regenerate this creature. Activate only if you control another
//	 colorless creature."
//
// "Can't block" is Rimrock Knight's restriction bit. The activation condition
// reads another colorless creature through the layers when the ability is
// activated; once on the stack the shield is made whether or not the creature
// stays.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ff3f5b1e-9d53-4258-84ee-cfecf685dd05",
		Name:            "Skitterskin",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordDevoid},
		Static:          []game.StaticAbility{effects.RestrictSelf(game.CantBlock)},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{B}: Regenerate this creature. Activate only if you control another colorless creature.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    effects.ManaCost("{1}{B}"),
			Condition: func(g *game.Game, controller, source uuid.UUID) bool {
				return effects.OldB308ControlsAnotherColorlessCreature(g, controller, source)
			},
			Effect: effects.RegenerateThisSource,
		}},
	})
}
