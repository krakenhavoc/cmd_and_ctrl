package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rustler Rampage — Instant {W} (Outlaws of Thunder Junction, #4468):
//
//	"Spree (Choose one or more additional costs.)
//	 + {1} — Untap all creatures target player controls.
//	 + {1} — Target creature gains double strike until end of turn."
//
// Two independently targeted, independently priced bullets (CR
// 702.172a), Requisition Raid's shape. The first untaps every tapped
// creature the chosen player controls as it resolves, snapshotted before
// the first untap (b16UntapAllYouControlMatching); the second is Boros
// Charm's double strike, declared as a benefit to its target. A bullet
// whose target has become illegal does nothing (CR 608.2b), and the other
// still happens.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ef0c0f28-e1ee-42a1-9878-95a0ceb0a9fa",
		Name:         "Rustler Rampage",
		Completeness: effects.CompletenessFull,
		Modes: effects.Spree(
			effects.SpreeModeDoing("Untap all creatures target player controls.", "{1}",
				effects.TargetPlayer("target player"),
				func(_ *game.StackItem, ctx *effects.Context, occ int) error {
					t, ok := effects.ModeTarget(ctx, occ)
					if !ok {
						return nil
					}
					return effects.B16UntapAllYouControlMatching(ctx, t.ID, game.Card.IsCreature)
				}),
			effects.ModeWithPurpose(effects.SpreeModeDoing("Target creature gains double strike until end of turn.", "{1}",
				effects.TargetCreature("target creature"),
				func(_ *game.StackItem, ctx *effects.Context, occ int) error {
					t, ok := effects.ModeTarget(ctx, occ)
					if !ok {
						return nil
					}
					return effects.GrantKeywordUntilEOT{
						Target:   t.ID,
						Keywords: []string{"double strike"},
						Label:    "Rustler Rampage — double strike",
					}.Apply(ctx)
				}), game.Purpose{Targets: game.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "double strike")), Answers: game.AnswerCombatGrant}),
		),
	})
}
