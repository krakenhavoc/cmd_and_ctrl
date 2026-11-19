package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Reckless Amplimancer — Creature — Elf Druid {1}{G}, 2/2 (Foundations
// Jumpstart):
//
//	"{4}{G}: Double this creature's power and toughness until end of turn."
//
// Doubling is +X/+Y with X and Y its power and toughness as the ability
// resolves (CR 701.10b), as Chameleon Colossus reads its power.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1c264f6e-caa7-44cb-8719-86212292679c",
		Name:         "Reckless Amplimancer",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{4}{G}: Double this creature's power and toughness until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    ManaCost("{4}{G}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !sourceIsStillThisPermanent(g, item) {
					return nil
				}
				return DoublePowerAndToughnessOfEachUntilEOT(NewContext(g, item), []uuid.UUID{item.SourceCardID},
					"Reckless Amplimancer — double power and toughness")
			},
		}},
	})
}
