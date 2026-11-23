package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "513034fc-873a-4877-afb2-39ece1ea6a7c",
		Name:         "Wingshield Agent",
		Completeness: effects.CompletenessFull,
		Replacements: effects.Snc6Entry("Wingshield Agent"),
		Triggered: []game.TriggeredAbility{
			effects.WithTargetsFrom(
				effects.WheneverThisAttacks("Wingshield Agent — up to one other target creature gains flying until end of turn",
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						id, ok := effects.B16FirstLegalTargetCard(ctx)
						if !ok {
							return nil
						}
						return effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"flying"}, Label: "Wingshield Agent — flying"}.Apply(ctx)
					}),
				effects.AnotherTarget(func(other effects.CardPredicate) *game.TargetSpec {
					return effects.UpToOneTargetCreature("up to one other target creature", other)
				})),
		},
	})
}
