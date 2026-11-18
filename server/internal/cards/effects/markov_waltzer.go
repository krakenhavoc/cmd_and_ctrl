package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "bc9ac394-15b6-4d95-acca-5dc247235d2b",
		Name:            "Markov Waltzer",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "haste"},
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(Targeting(
				AtBeginningOfYourCombat(label, func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						if t.Kind != game.TargetCard {
							continue
						}
						if err := untilEndOfTurn(ctx, t.ID, nil, label, game.ModifyPTMod(1, 0)); err != nil {
							return err
						}
					}
					return nil
				}),
				TargetCreature("up to two target creatures you control", YouControl()).WithCount(0, 2)),
				ForTargets(BenefitUntilEOT(0, 1, 0))),
		},
	})
}
