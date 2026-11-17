package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "41d408dc-f6fe-45b2-b225-2ac25db9ad0f",
		Name:         "Perrie, the Pulverizer",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Perrie — put a shield counter on target creature", func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				id, ok := b16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				return AddCounter{Target: id, Kind: game.CounterShield, N: 1}.Apply(ctx)
			}), TargetCreature("target creature")),
			Targeting(WheneverThisAttacks("Perrie — target creature you control gains trample and gets +X/+X", func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				id, ok := b16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				x := len(sncB07KindsAmongPermanentsYouControl(g, item.Controller))
				if err := (GrantKeywordUntilEOT{Target: id, Keywords: []string{"trample"}, Label: "Perrie — trample"}).Apply(ctx); err != nil {
					return err
				}
				return BoostUntilEOT{Target: id, Power: x, Toughness: x, Label: "Perrie — +X/+X"}.Apply(ctx)
			}), TargetCreature("target creature you control", YouControl())),
		},
	})
}
