package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fevered Strength — Instant {2}{B} (Alliances, #5634):
//
//	"Target creature gets +2/+0 until end of turn.
//	 Draw a card at the beginning of the next turn's upkeep."
//
// A layer 7c boost (CR 613.4c) that ends at cleanup (CR 514.2), then the
// delayed draw at the next upkeep at the table, whoever's turn it is
// (CR 603.7a, next_upkeep_draw.go). With its one target gone the spell
// does not resolve at all, so nothing is drawn either (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "52512d7a-9b1c-4475-80aa-c95f4efa1015",
		Name:         "Fevered Strength",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		Purpose:      ForTargets(BenefitUntilEOT(0, 2, 0)),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if err := boostTheTargetUntilEOT(2, 0, "Fevered Strength — +2/+0 until end of turn")(item, ctx); err != nil {
				return err
			}
			return drawACardAtTheNextUpkeep("Fevered Strength — draw a card").Apply(ctx)
		},
	})
}
