package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fang of the Pack — Creature — Wolf {5}{G}, 5/3 (Conspiracy: Take the
// Crown, S82 older sets batch 270, #5540):
//
//	"Melee (Whenever this creature attacks, it gets +1/+1 until end of turn
//	 for each opponent you attacked this combat.)
//	 At the beginning of combat on your turn, another target creature you
//	 control gains melee until end of turn. (If a creature has multiple
//	 instances of melee, each triggers separately.)"
//
// Duke Ulder Ravengard's trigger shape (another target creature you
// control, at the beginning of combat on your turn) with melee as the
// keyword granted until end of turn. Melee is cumulative (CR 702.121b), so
// a target that already has melee keeps both instances and each triggers
// on its own attack.
//
// No simplification.
func init() {
	const label = "Fang of the Pack — another target creature you control gains melee until end of turn"
	effects.Register(effects.Spec{
		OracleID:        "ae8dffde-e59f-4f4f-8e7f-e7dc34f62e92",
		Name:            "Fang of the Pack",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordMelee},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.On(game.EventStepBegan, effects.StepBegan(game.StepBeginCombat, true), label,
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						for _, t := range ctx.LegalTargets() {
							if t.Kind != game.TargetCard {
								continue
							}
							return effects.GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{game.KeywordMelee}, Label: label}.Apply(ctx)
						}
						return nil
					}),
				effects.Another(effects.TargetCreature("another target creature you control", effects.YouControl()))),
		},
	})
}
