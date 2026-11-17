package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "18da85d4-2a51-45c0-a18c-bce9bc39fa64",
		Name:            "Gnarlroot Pallbearer",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Gnarlroot Pallbearer — target creature gets +X/+X until end of turn, where X is the number of creature cards in your graveyard",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					id, ok := b16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					x := b11CreatureCardsInGraveyard(g, item.Controller)
					if x <= 0 {
						return nil
					}
					return BoostUntilEOT{Target: id, Power: x, Toughness: x, Label: "Gnarlroot Pallbearer — +X/+X"}.Apply(ctx)
				}),
				TargetCreature("target creature")),
		},
	})
}
