package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Prismabasher — Creature — Elemental {4}{G}{G} (Lorwyn Eclipsed, #4458):
//
//	"Trample
//	 Vivid — When this creature enters, up to X target creatures you control
//	 get +X/+X until end of turn, where X is the number of colors among
//	 permanents you control."
//
// X sets how many targets can be chosen (counted as the trigger is put on
// the stack, the Prismabasher already on the battlefield) and the size of
// the boost (counted as it resolves).
func init() {
	t := WhenThisEnters("Prismabasher — up to X target creatures you control get +X/+X until end of turn (vivid)",
		prismabasherPump)
	// X is counted once, as the trigger triggers, and carried on its context.
	t.StampColorsAmongYours = true
	t.TargetsFrom = func(tc game.TriggerContext, _ *game.Card, _ *game.Game) *game.TargetSpec {
		if tc.ColorsAmongYours <= 0 {
			return nil
		}
		return TargetCreature("up to X target creatures you control", YouControl()).WithCount(0, tc.ColorsAmongYours)
	}
	Register(Spec{
		OracleID:        "f75be4f2-d1a7-40c7-a0e4-ba8b03556ac6",
		Name:            "Prismabasher",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered:       []game.TriggeredAbility{t},
	})
}

func prismabasherPump(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	x := len(colorsAmongPermanentsYouControl(g, ctx.Controller()))
	if x == 0 {
		return nil
	}
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		if err := (BoostUntilEOT{Target: t.ID, Power: x, Toughness: x,
			Label: "Prismabasher — +X/+X until end of turn"}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}
