package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Subterranean Spirit — Creature — Elemental Spirit {3}{R}{R}, 3/3 (Mirage,
// #5633):
//
//	"Protection from red
//	 {T}: This creature deals 1 damage to each creature without flying."
//
// It damages itself too (it has no flying); the Spirit is red, so protection
// from red prevents that damage to itself (CR 702.16e).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "58917163-584d-4c74-8243-f56c5d4b49e7",
		Name:            "Subterranean Spirit",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"protection from red"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: This creature deals 1 damage to each creature without flying.",
			Cost:    effects.TapCost(),
			Purpose: game.Purpose{Answers: game.AnswerRemove, Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 1, Partial: true}},
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.DamageEachMatching(effects.NewContext(g, item), effects.And(effects.Creature(), effects.WithoutKeyword("flying")), 1)
			},
		}},
	})
}
