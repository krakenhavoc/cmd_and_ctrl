package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Thopter Architect — Creature — Human Artificer {3}{W}, 2/3 (The Brothers'
// War, #5482):
//
//	"Whenever an artifact you control enters, target creature gains flying
//	 until end of turn."
//
// Any creature is a legal target, not only yours; it is chosen as the
// trigger goes on the stack and checked again as it resolves (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3194fe3d-57d3-4dab-ab39-3778b75150fa",
		Name:         "Thopter Architect",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(Targeting(
				On(game.EventETB, ArtifactEnteredUnderYourControl,
					"Thopter Architect — target creature gains flying until end of turn",
					func(g *game.Game, item *game.StackItem) error {
						ctx := NewContext(g, item)
						id, ok := b16FirstLegalTargetCard(ctx)
						if !ok {
							return nil
						}
						return GrantKeywordUntilEOT{Target: id, Keywords: []string{"flying"}, Label: "Thopter Architect"}.Apply(ctx)
					}),
				TargetCreature("target creature")),
				ForTargets(BenefitUntilEOT(0, 0, 0, "flying"))),
		},
	})
}
