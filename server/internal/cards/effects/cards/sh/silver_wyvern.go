package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Silver Wyvern — Creature — Drake {3}{U}{U}, 4/3 (Stronghold, S82 OLD batch
// 746, #5627):
//
//	"Flying
//	 {U}: Change the target of target spell or ability that targets only this
//	 creature. The new target must be a creature."
//
// The clause is any spell or ability whose targets are all this creature
// (game.TargetSpec.ItemTargetsOnlySource), judged as it is announced and again
// as the ability resolves (CR 608.2b). The retarget is Bolt Bend's
// RetargetChangeOne with the destination limited to creatures on the
// battlefield (RetargetOffer.NewTargetCreature) on top of the original clause's
// own rules: the new target is legal for the spell or ability (CR 115.7a) and
// the controller of the Wyvern's ability chooses it; with no creature to move
// to, the target is unchanged.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "4baa8f05-1ef8-4770-8c66-ef5f0f02a717",
		Name:            "Silver Wyvern",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{U}: Change the target of target spell or ability that targets only this creature. The new target must be a creature.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    effects.ManaCost("{U}"),
			Targets: effects.TargetsOnlyThisCreature(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				for _, t := range item.Targets {
					if t.Kind != game.TargetCard || !ctx.IsTargetLegal(t) {
						continue
					}
					return effects.ChangeTargets{
						StackID:           t.ID,
						Policy:            game.RetargetChangeOne,
						Reason:            "Silver Wyvern — choose a new target (a creature)",
						NewTargetCreature: true,
					}.Apply(ctx)
				}
				return nil
			},
		}},
	})
}
