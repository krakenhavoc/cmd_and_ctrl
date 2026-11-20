package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Warhorn Blast — Instant {4}{W} (Kaldheim, #5505):
//
//	"Creatures you control get +2/+1 until end of turn.
//	 Foretell {2}{W}"
//
// Inspired Charge's team pump, plus the shared foretell special action
// (CR 702.143). The set of creatures is locked in as it resolves
// (CR 611.2c), so one that enters later gets nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:       "8106606c-9a1c-44b2-a9bf-3f189ce050ef",
		Name:           "Warhorn Blast",
		Completeness:   CompletenessFull,
		SpecialActions: []game.SpecialAction{Foretell("{2}{W}")},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return BoostUntilEOT{
				Match:     And(Creature(), YouControl()),
				Power:     2,
				Toughness: 1,
				Label:     "Warhorn Blast — +2/+1",
			}.Apply(ctx)
		},
	})
}
