package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Urabrask's Anointer — Artifact Creature — Phyrexian Wizard {3}{R}, 4/2
// (Phyrexia: All Will Be One):
//
//	"When this creature enters, it deals X damage to any target, where X
//	 is the number of permanents you control with oil counters on them."
//
// X is counted as the ability resolves (CR 608.2), over permanents you
// control with one or more oil counters, any type.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a09c134b-7802-4526-9250-ec5cc637d9c0",
		Name:         "Urabrask's Anointer",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Urabrask's Anointer — it deals X damage to any target, where X is the number of permanents you control with oil counters on them",
					func(g *game.Game, item *game.StackItem) error {
						ctx := NewContext(g, item)
						for _, t := range ctx.LegalTargets() {
							return DealDamage{Source: item.SourceCardID, Target: t.ID, Amount: s82OilPermanentsYouControl(ctx)}.Apply(ctx)
						}
						return nil
					}),
				TargetAny()),
		},
	})
}
