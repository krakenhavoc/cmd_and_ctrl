package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Aerial Boost — Instant {1}{W} (March of the Machine): "Convoke. Target
// creature gets +2/+2 and gains flying until end of turn."
//
// Acrobatic Leap's shape: a boost and a keyword grant on the one target.
func init() {
	Register(Spec{
		OracleID:     "d7fa5c42-7cd5-42e7-a42e-724c1a3df96a",
		Name:         "Aerial Boost",
		Completeness: CompletenessFull,
		TapCost:      Convoke(),
		Purpose:      ForTargets(BenefitUntilEOT(0, 2, 2, "flying")),
		Targets:      TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			for _, id := range legalTargetCardIDs(ctx) {
				if err := (BoostUntilEOT{Target: id, Power: 2, Toughness: 2, Label: "Aerial Boost — +2/+2"}).Apply(ctx); err != nil {
					return err
				}
				if err := (GrantKeywordUntilEOT{Target: id, Keywords: []string{"flying"}, Label: "Aerial Boost — flying"}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
