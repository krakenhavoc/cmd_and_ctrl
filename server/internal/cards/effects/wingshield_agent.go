package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wingshield Agent — Creature — Human Soldier {2}{U}, 2/3 (Streets of
// New Capenna):
//
//	"This creature enters with a shield counter on it. (If it would be
//	 dealt damage or destroyed, remove a shield counter from it
//	 instead.)
//	 Whenever this creature attacks, up to one other target creature
//	 gains flying until end of turn."
//
// "Other" keeps the Agent out of its own target list; with no target
// the trigger does nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "513034fc-873a-4877-afb2-39ece1ea6a7c",
		Name:         "Wingshield Agent",
		Completeness: CompletenessFull,
		Replacements: snc6Entry("Wingshield Agent"),
		Triggered: []game.TriggeredAbility{
			withTargetsFrom(
				WheneverThisAttacks("Wingshield Agent — up to one other target creature gains flying until end of turn",
					func(g *game.Game, item *game.StackItem) error {
						ctx := NewContext(g, item)
						id, ok := b16FirstLegalTargetCard(ctx)
						if !ok {
							return nil
						}
						return GrantKeywordUntilEOT{Target: id, Keywords: []string{"flying"}, Label: "Wingshield Agent — flying"}.Apply(ctx)
					}),
				AnotherTarget(func(other CardPredicate) *game.TargetSpec {
					return UpToOneTargetCreature("up to one other target creature", other)
				})),
		},
	})
}
