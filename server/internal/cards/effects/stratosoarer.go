package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Stratosoarer — Creature — Elemental {4}{U}, 3/5 (Lorwyn Eclipsed,
// #4458):
//
//	"Flying
//	 When this creature enters, target creature gains flying until end of
//	 turn.
//	 Basic landcycling {1}{U}"
//
// The enter trigger grants flying to any creature, the Stratosoarer
// included, until end of turn; a target that left is skipped
// (CR 608.2b). Basic landcycling is the cycling family's BasicLandcycling.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "36ab41b7-bcfb-4618-bbfb-70d608dba15a",
		Name:            "Stratosoarer",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Stratosoarer — target creature gains flying until end of turn", stratosoarerGrantFlying),
				TargetCreature("target creature")),
		},
		Activated: []ActivatedAbility{BasicLandcycling("{1}{U}")},
	})
}

func stratosoarerGrantFlying(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	return GrantKeywordUntilEOT{Target: id, Keywords: []string{"flying"},
		Label: "Stratosoarer — gains flying until end of turn"}.Apply(ctx)
}
