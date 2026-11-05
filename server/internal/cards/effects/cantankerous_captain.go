package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cantankerous Captain — {2}{R} Creature — Tellarite Officer, 2/3 (Star
// Trek, #4973):
//
//	"At the beginning of combat on your turn, target creature you control
//	 gets +1/+0 until end of turn. Then that creature gains menace until
//	 end of turn if creatures you control have total power 8 or greater.
//	 (It can't be blocked except by two or more creatures.)"
//
// Battle-Rattle Shaman's targeted combat trigger. "Then" orders the two
// halves: the +1/+0 lands first, so it counts toward the total, and the
// total is read as the ability resolves over every creature its
// controller controls then, after the layers (b43TotalPowerControlled).
// A target that left in response gets nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2375e6e4-bea2-4fa2-b0d5-467e3076962c",
		Name:         "Cantankerous Captain",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(Targeting(
				AtBeginningOfYourCombat("Cantankerous Captain — target creature you control gets +1/+0, then menace if your creatures have total power 8 or greater",
					cantankerousCaptainCombat),
				TargetCreature("target creature you control", YouControl())),
				ForTargets(BenefitUntilEOT(0, 1, 0))),
		},
	})
}

func cantankerousCaptainCombat(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	if err := (BoostUntilEOT{Target: id, Power: 1, Label: "Cantankerous Captain — +1/+0"}).Apply(ctx); err != nil {
		return err
	}
	if b43TotalPowerControlled(g, item.Controller) < 8 {
		return nil
	}
	return GrantKeywordUntilEOT{
		Target:   id,
		Keywords: []string{"menace"},
		Label:    "Cantankerous Captain — menace until end of turn",
	}.Apply(ctx)
}
