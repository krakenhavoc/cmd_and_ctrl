package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ghirapur Guide — Creature — Elf Scout {2}{G}, 3/2:
//
//	"{2}{G}: Target creature you control can't be blocked by creatures with
//	 power 2 or less this turn."
//
// Rhonas's Stalwart's block rule on a chosen creature: a scoped effect pinned
// to the target at resolution (CR 611.2c) and swept at end of turn, the
// blocker's power read live as blockers are declared (CR 509.1b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "92d5bb7e-08ba-448e-8763-f47c89505e80",
		Name:         "Ghirapur Guide",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{G}: Target creature you control can't be blocked by creatures with power 2 or less this turn.",
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant, Targets: game.ForTargets(game.TargetPurpose{Slot: 0, UntilEndOfTurn: game.TargetBenefit})},
			Cost:    effects.ManaCost("{2}{G}"),
			Targets: effects.TargetCreature("target creature you control", effects.YouControl()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				id, ok := effects.B16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				return effects.CantBeBlockedThisTurnByPower{Target: id, AtMost: 2, Label: "Ghirapur Guide — can't be blocked by creatures with power 2 or less"}.Apply(ctx)
			},
		}},
	})
}
