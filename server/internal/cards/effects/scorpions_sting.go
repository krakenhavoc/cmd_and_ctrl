package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Scorpion's Sting — Instant {1}{B} (Marvel's Spider-Man):
//
//	"Target creature gets -3/-3 until end of turn."
//
// One target clause; a target that is gone by resolution counters the spell
// (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7f5e5aa8-3f15-45c9-bfc8-1f4b98d2be38",
		Name:         "Scorpion's Sting",
		Purpose:      ForTargets(HarmUntilEOT(0, -3, -3)),
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			return BoostUntilEOT{
				Target:    item.Targets[0].ID,
				Power:     -3,
				Toughness: -3,
				Label:     "Scorpion's Sting — -3/-3 until end of turn",
			}.Apply(ctx)
		},
	})
}
