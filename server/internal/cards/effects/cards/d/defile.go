package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Defile — Instant {B}:
//
//	"Target creature gets -1/-1 until end of turn for each Swamp you
//	 control."
//
// The Swamp count is taken once, as the spell resolves (CR 611.2c: a
// continuous effect from a resolving spell locks in its set and its
// amount), so a Swamp played afterwards does not shrink the creature
// further. With no Swamps the spell does nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "49dddec0-d958-4810-8c6e-225fc8118c8f",
		Name:         "Defile",
		Purpose:      effects.ForTargets(effects.HarmUntilEOT(0, 0, 0)),
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			n := effects.SwampsControlledBy(ctx.Game, ctx.Controller(), ctx.TextSubtype("Swamp"))
			return effects.BoostUntilEOT{
				Target:    item.Targets[0].ID,
				Power:     -n,
				Toughness: -n,
				Label:     "Defile — -1/-1 for each Swamp you control",
			}.Apply(ctx)
		},
	})
}
