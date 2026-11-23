package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Markov Waltzer — Creature — Vampire {2}{R}{W}, 1/3 (Innistrad: Crimson
// Vow, #5495):
//
//	"Flying, haste
//	 At the beginning of combat on your turn, up to two target creatures
//	 you control each get +1/+0 until end of turn."
//
// Flying and haste are printed keywords. The trigger is Earth Kingdom
// Soldier's "up to two target creatures you control" on Wild Pack Squad's
// beginning-of-combat shape: targets are chosen as it goes on the stack, and
// each one still legal as it resolves gets +1/+0 until end of turn
// (CR 608.2b).
//
// No simplification.
func init() {
	const label = "Markov Waltzer — up to two target creatures you control each get +1/+0 until end of turn"
	effects.Register(effects.Spec{
		OracleID:        "bc9ac394-15b6-4d95-acca-5dc247235d2b",
		Name:            "Markov Waltzer",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "haste"},
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.Targeting(
				effects.AtBeginningOfYourCombat(label, func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						if t.Kind != game.TargetCard {
							continue
						}
						if err := effects.UntilEndOfTurn(ctx, t.ID, nil, label, game.ModifyPTMod(1, 0)); err != nil {
							return err
						}
					}
					return nil
				}),
				effects.TargetCreature("up to two target creatures you control", effects.YouControl()).WithCount(0, 2)),
				effects.ForTargets(effects.BenefitUntilEOT(0, 1, 0))),
		},
	})
}
