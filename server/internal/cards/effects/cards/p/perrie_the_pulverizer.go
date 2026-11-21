package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Perrie, the Pulverizer — Legendary Creature — Rhino Soldier {1}{G}{W}{U},
// 3/3 (New Capenna Commander, #5491):
//
//	"When Perrie enters, put a shield counter on target creature. (If it
//	 would be dealt damage or destroyed, remove a shield counter from it
//	 instead.)
//	 Whenever Perrie attacks, target creature you control gains trample
//	 and gets +X/+X until end of turn, where X is the number of different
//	 kinds of counters among permanents you control."
//
// X is read as the attack trigger resolves (CR 608.2h), over every
// permanent its controller has: a +1/+1 and a shield counter on one
// creature and a shield counter on another are two kinds. A target that
// has left by then makes the ability do nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "41d408dc-f6fe-45b2-b225-2ac25db9ad0f",
		Name:         "Perrie, the Pulverizer",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Perrie — put a shield counter on target creature", effects.ShieldCounterOnTarget), effects.TargetCreature("target creature")),
			effects.Targeting(effects.WheneverThisAttacks("Perrie — target creature you control gains trample and gets +X/+X", func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				id, ok := effects.B16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				x := len(effects.SncB07KindsAmongPermanentsYouControl(g, item.Controller))
				if err := (effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"trample"}, Label: "Perrie — trample"}).Apply(ctx); err != nil {
					return err
				}
				return effects.BoostUntilEOT{Target: id, Power: x, Toughness: x, Label: "Perrie — +X/+X"}.Apply(ctx)
			}), effects.TargetCreature("target creature you control", effects.YouControl())),
		},
	})
}
