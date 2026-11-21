package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gale Swooper — Creature — Griffin {3}{W}, 3/2 (Core Set 2021, #5511):
//
//	"Flying
//	 When this creature enters, target creature gains flying until end of
//	 turn."
//
// Stratosoarer's enters trigger: a flying grant on the target that ends
// at cleanup; a target that left in response gets nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "1758e102-22da-4cb3-98d5-76f7501d026d",
		Name:            "Gale Swooper",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Gale Swooper — target creature gains flying until end of turn", galeSwooperGrantFlying),
				effects.TargetCreature("target creature")),
		},
	})
}

func galeSwooperGrantFlying(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	id, ok := effects.B16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	return effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"flying"},
		Label: "Gale Swooper — gains flying until end of turn"}.Apply(ctx)
}
