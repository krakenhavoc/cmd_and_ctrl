package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Glister Bairn — Creature — Ouphe {2}{G/U}{G/U}{G/U}, 1/4 (Lorwyn
// Eclipsed, #4458):
//
//	"Vivid — At the beginning of combat on your turn, another target
//	 creature you control gets +X/+X until end of turn, where X is the
//	 number of colors among permanents you control."
//
// Agent Bishop's beginning-of-combat trigger with Little Bear's "another
// target creature you control" clause. X is counted as the trigger
// resolves (CR 608.2h), through Faeburrow Elder's
// colorsAmongPermanentsYouControl: the union of every permanent's current
// colours, so the green-and-blue Bairn itself counts two and a colourless
// permanent counts none. The pump is locked in once it resolves, so a
// colour gained or lost later in the turn does not change it. Vivid is an
// ability word with no rules meaning of its own (CR 207.2c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6a7229a8-1634-4278-ab50-0ec2b51edac1",
		Name:         "Glister Bairn",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				AtBeginningOfYourCombat("Glister Bairn — another target creature you control gets +X/+X, X the colors among permanents you control",
					glisterBairnPump),
				Another(TargetCreature("another target creature you control", YouControl()))),
		},
	})
}

func glisterBairnPump(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	x := len(colorsAmongPermanentsYouControl(g, ctx.Controller()))
	return BoostUntilEOT{Target: id, Power: x, Toughness: x,
		Label: "Glister Bairn — +X/+X until end of turn"}.Apply(ctx)
}
