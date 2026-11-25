package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Benevolent Bodyguard — {W} Creature — Human Cleric 1/1 (Eternal Masters,
// S82 OLD, #5542):
//
//	"Sacrifice this creature: Target creature you control gains protection
//	 from the color of your choice until end of turn."
//
// Mother of Runes' effect with a sacrifice cost: the target is chosen as the
// ability is announced and the color as it resolves, so an opponent who
// responds commits first. The Bodyguard is sacrificed as a cost, so it is not a
// legal target in practice.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c8c38570-0b42-4bfc-896f-f57f9a11ba9d",
		Name:         "Benevolent Bodyguard",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "Sacrifice this creature: Target creature you control gains protection from the color of your choice until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerProtect, Targets: game.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "protection"))},
			Cost:    effects.SacrificeThis(),
			Targets: effects.TargetCreature("target creature you control", effects.YouControl()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.ProtectionFromAChosenColorForTheTarget("Benevolent Bodyguard")(item, effects.NewContext(g, item))
			},
		}},
	})
}
