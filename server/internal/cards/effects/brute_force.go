package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Brute Force — Instant {R} (EDHREC rank 4429):
//
//	"Target creature gets +3/+3 until end of turn."
//
// Giant Growth in red, and nothing else. It is in roadmap batch 42
// (#449) because the roadmap filed it under until-end-of-turn
// continuous effects, which #279 shipped in S32: the turn-scoped
// layer-7 static that BoostUntilEOT registers is now the whole card.
//
// The pump is a Layer 7c modification pinned to the one target
// chosen at announce and re-checked at resolution (CR 608.2b), so a
// Brute Force whose target left the battlefield in response does
// nothing at all rather than drifting onto another creature.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9880ba09-d5b8-4675-bfb4-2161d86d2d41",
		Name:         "Brute Force",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			return BoostUntilEOT{
				Target: item.Targets[0].ID,
				Power:  3, Toughness: 3,
				Label: "Brute Force — +3/+3 until end of turn",
			}.Apply(ctx)
		},
	})
}
