package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Archon of Justice — Creature — Archon {3}{W}{W}, 4/4 (Jumpstart):
//
//	"Flying
//	 When this creature dies, exile target permanent."
//
// A dies trigger (CR 603.10a looks back in time, so it fires from the
// graveyard) whose target is any permanent, a land or one of its
// controller's own included, and is mandatory: with any permanent on the
// battlefield one must be chosen. The target is re-checked on resolution
// (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "d551931d-4ebe-4acc-9e87-6c628897071a",
		Name:            "Archon of Justice",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisDies("Archon of Justice — exile target permanent", func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				id, ok := b16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				return ExileTarget{Target: id}.Apply(ctx)
			}), TargetPermanent("target permanent")),
		},
	})
}
