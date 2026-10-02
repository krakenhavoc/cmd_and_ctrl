package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Shimmering Mirage — Instant {1}{U}:
//
//	"Target land becomes the basic land type of your choice until end of
//	 turn.
//	 Draw a card."
//
// ADR 0109 §1 (#1881): CR 305.7 from a resolved spell. The type is chosen
// as it resolves (CR 608.2); until end of turn the land's land types are
// replaced by it (its other subtypes stay, CR 205.1a), it loses the
// abilities its rules text gives it and taps for the chosen colour (CR
// 305.6). The draw comes after the choice, in the order printed (CR
// 608.2c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d1720ddc-d6ad-4d78-8dd7-29539bbb73da",
		Name:         "Shimmering Mirage",
		Completeness: CompletenessFull,
		Targets:      TargetPermanent("target land", Land()),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return TargetLandBecomesChosenTypeThen(ctx, "Shimmering Mirage", DrawCards{N: 1}.Apply)
		},
	})
}
