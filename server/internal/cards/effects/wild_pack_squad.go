package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "294b2aa7-cb46-46bd-8241-69b26d73aa2d",
		Name:         "Wild Pack Squad",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(Targeting(
				AtBeginningOfYourCombat(label, func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					id, ok := b16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					return GrantKeywordUntilEOT{Target: id, Keywords: []string{"first strike", "vigilance"}, Label: label}.Apply(ctx)
				}),
				TargetCreature("up to one target creature").WithCount(0, 1)),
				ForTargets(BenefitUntilEOT(0, 0, 0, "first strike", "vigilance"))),
		},
	})
}
