package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:        "36ab41b7-bcfb-4618-bbfb-70d608dba15a",
		Name:            "Stratosoarer",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Stratosoarer — target creature gains flying until end of turn", stratosoarerGrantFlying),
				effects.TargetCreature("target creature")),
		},
		Activated: []effects.ActivatedAbility{effects.BasicLandcycling("{1}{U}")},
	})
}

func stratosoarerGrantFlying(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	id, ok := effects.B16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	return effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"flying"},
		Label: "Stratosoarer — gains flying until end of turn"}.Apply(ctx)
}
