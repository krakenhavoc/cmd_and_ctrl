package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sudden Death — Instant {1}{B}{B}:
//
//	"Split second (As long as this spell is on the stack, players
//	 can't cast spells or activate abilities that aren't mana
//	 abilities.)
//	 Target creature gets -4/-4 until end of turn."
//
// A layer-7c shrink (BoostUntilEOT with negative values) behind split
// second (#1519), so the creature's controller cannot pump it or
// sacrifice it for value in response. A creature that is indestructible
// still dies — the 0-toughness state-based action is not destruction
// (CR 704.5f), and the shrink reaches it through the ordinary SBA.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "a3475ca6-d88e-4855-a554-1338624d1635",
		Name:            "Sudden Death",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordSplitSecond},
		Targets:         TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			return BoostUntilEOT{
				Target:    item.Targets[0].ID,
				Power:     -4,
				Toughness: -4,
				Label:     "Sudden Death — -4/-4 until end of turn",
			}.Apply(ctx)
		},
	})
}
