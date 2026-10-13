package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Overkill — Instant {2}{B}:
//
//	"Target creature gets -0/-9999 until end of turn."
//
// A Layer 7c toughness modification, so the creature dies to the
// zero-toughness state-based action (CR 704.5f) rather than to
// "destroy": indestructible does not save it. (#3213, the Kuja deck.)
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "571cb4b6-b4e5-40a2-a970-42c1ab65c8a9",
		Name:         "Overkill",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			return BoostUntilEOT{
				Target:    item.Targets[0].ID,
				Toughness: -9999,
				Label:     "Overkill — -0/-9999 until end of turn",
			}.Apply(ctx)
		},
	})
}
