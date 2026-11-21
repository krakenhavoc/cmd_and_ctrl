package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gnarlroot Pallbearer — Creature — Treefolk Druid {4}{G}{G}, 5/5
// (The Brothers' War):
//
//	"Trample
//	 When this creature enters, target creature gets +X/+X until end of
//	 turn, where X is the number of creature cards in your graveyard."
//
// Daybreak Charger's targeted enters pump with X counted as the ability
// resolves (CR 608.2h). With no creature card in the graveyard the pump is
// +0/+0 and does nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "18da85d4-2a51-45c0-a18c-bce9bc39fa64",
		Name:            "Gnarlroot Pallbearer",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Gnarlroot Pallbearer — target creature gets +X/+X until end of turn, where X is the number of creature cards in your graveyard",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					id, ok := effects.B16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					x := effects.B11CreatureCardsInGraveyard(g, item.Controller)
					if x <= 0 {
						return nil
					}
					return effects.BoostUntilEOT{Target: id, Power: x, Toughness: x, Label: "Gnarlroot Pallbearer — +X/+X"}.Apply(ctx)
				}),
				effects.TargetCreature("target creature")),
		},
	})
}
