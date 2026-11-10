package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Impolite Entrance — Sorcery {R} (Lorwyn Eclipsed, #4458):
//
//	"Target creature gains trample and haste until end of turn.
//	 Draw a card."
//
// The keywords are a turn-long grant on the target as it is now, then the
// draw. With the target illegal as the spell resolves nothing happens, and
// no card is drawn (the ruling, CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b77764fe-bc79-410a-995a-ac392f5b46b8",
		Name:         "Impolite Entrance",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		Purpose:      ForTargets(BenefitUntilEOT(0, 0, 0, "trample", "haste")),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			id, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			if err := (GrantKeywordUntilEOT{Target: id, Keywords: []string{"trample", "haste"},
				Label: "Impolite Entrance — trample and haste"}).Apply(ctx); err != nil {
				return err
			}
			return DrawCards{N: 1}.Apply(ctx)
		},
	})
}
