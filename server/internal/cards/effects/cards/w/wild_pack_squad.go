package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wild Pack Squad — Creature — Human Mercenary {2}{W}, 2/3 (Marvel Super
// Heroes Commander, #4455):
//
//	"At the beginning of combat on your turn, up to one target creature
//	 gains first strike and vigilance until end of turn."
//
// Eccentric Apprentice's beginning-of-combat "up to one target creature"
// trigger with Origin of Captain America's grant. No target is legal
// (CR 115.1b); a target that left by resolution gets nothing (CR 608.2b).
//
// No simplification.
func init() {
	const label = "Wild Pack Squad — up to one target creature gains first strike and vigilance until end of turn"
	effects.Register(effects.Spec{
		OracleID:     "294b2aa7-cb46-46bd-8241-69b26d73aa2d",
		Name:         "Wild Pack Squad",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.Targeting(
				effects.AtBeginningOfYourCombat(label, func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					id, ok := effects.B16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					return effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"first strike", "vigilance"}, Label: label}.Apply(ctx)
				}),
				effects.TargetCreature("up to one target creature").WithCount(0, 1)),
				effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "first strike", "vigilance"))),
		},
	})
}
