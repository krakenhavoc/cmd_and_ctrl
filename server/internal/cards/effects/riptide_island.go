package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Riptide Island — Plane — Dominaria (Commander Masters, #4870):
//
//	"When you planeswalk to Riptide Island and at the beginning of your
//	 upkeep, create two 1/1 colorless Sliver creature tokens.
//	 Whenever chaos ensues, Slivers you control gain haste and get +X/+X
//	 until end of turn, where X is the number of Slivers you control."
//
// A plane (ADR 0216 §6): one ability with two trigger conditions (the
// planeswalk and the planar controller's upkeep) makes two Slivers. The
// chaos ability (CR 311.7) counts the planar controller's Slivers as it
// resolves (the ruling), and the Slivers it affects are the ones it counts
// (CR 611.2c): a Sliver made later is not pumped.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "91fe554b-13db-451d-ab25-717ab2b70649",
		Name:         "Riptide Island",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenYouPlaneswalkToThisAndAtYourUpkeep("Riptide Island — create two 1/1 colorless Sliver creature tokens",
				Do(CreateToken{Template: TokenCard("1/1 colorless Sliver"), N: 2})),
			WheneverChaosEnsues("Riptide Island — Slivers you control gain haste and get +X/+X until end of turn, where X is the number of Slivers you control",
				riptideIslandChaos),
		},
	})
}

func riptideIslandChaos(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	slivers := And(Creature(), YouControl(), Subtype("Sliver"))
	x := len(MatchingBattlefield(ctx, slivers))
	if x == 0 {
		return nil
	}
	if err := (GrantKeywordUntilEOT{Match: slivers, Keywords: []string{"haste"}, Label: "Riptide Island — haste"}).Apply(ctx); err != nil {
		return err
	}
	return BoostUntilEOT{Match: slivers, Power: x, Toughness: x, Label: "Riptide Island — +X/+X"}.Apply(ctx)
}
