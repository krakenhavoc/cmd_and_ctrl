package cardslu

import (
	"strings"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mage Duel — Sorcery {2}{G} (Strixhaven: School of Mages, #5502):
//
//	"This spell costs {2} less to cast if you've cast another instant or
//	 sorcery spell this turn.
//	 Target creature you control gets +1/+2 until end of turn. Then it fights
//	 target creature you don't control."
//
// The reduction reads the caster's cast tally as the cost is determined
// (CR 601.2f): the spell being cast is not in the tally yet (it becomes cast
// at CR 601.2i), so a tallied instant or sorcery is another one. The body is
// Pit Fight's two-clause fight (slot zero fights slot one, nothing fights once
// either left or became illegal, CR 701.14b) after the +1/+2, which the first
// target gets if it is still legal.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "763882f9-a7ad-4fb1-bdf4-c9a158d4b68e",
		Name:         "Mage Duel",
		Completeness: effects.CompletenessFull,
		Purpose: game.Purpose{Answers: game.AnswerRemove, // #4734: a fight
			Targets: game.ForTargets(effects.BenefitUntilEOT(0, 1, 2))},
		SelfCostModifiers: []game.CostModifier{
			effects.CostsLess(2, "This spell costs {2} less to cast if you've cast another instant or sorcery spell this turn.",
				func(q game.CostQuery) bool {
					if q.Game == nil {
						return false
					}
					types := " " + q.Game.CastTallyFor(q.Controller).CardTypes + " "
					return strings.Contains(types, " Instant ") || strings.Contains(types, " Sorcery ")
				}),
		},
		Targets: effects.Clauses(
			effects.TargetCreature("target creature you control", effects.YouControl()),
			effects.TargetCreature("target creature you don't control", effects.OpponentControls()),
		),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if mine, ok := ctx.ClauseTarget(0); ok && mine.Kind == game.TargetCard {
				if err := (effects.BoostUntilEOT{Target: mine.ID, Power: 1, Toughness: 2, Label: "Mage Duel — +1/+2"}).Apply(ctx); err != nil {
					return err
				}
			}
			return effects.SlotZeroFightsSlotOne(ctx.Game, item)
		},
	})
}
