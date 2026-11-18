package effects

import (
	"slices"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ray of Enfeeblement — Instant {B} (Adventures in the Forgotten Realms):
//
//	"Target creature gets -4/-1 until end of turn. If that creature is white,
//	 it gets -4/-4 until end of turn instead."
//
// The colour is read through the layered view as the spell resolves; a
// creature that is white gets -4/-4 and not both.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "647e5639-6339-48af-af96-e48729e1cad3",
		Name:         "Ray of Enfeeblement",
		Completeness: CompletenessFull,
		Purpose:      ForTargets(HarmUntilEOT(0, -4, -1)),
		Targets:      TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			id := item.Targets[0].ID
			toughness := -1
			if c, ok := ctx.Game.LookupCardForEffect(id); ok && slices.Contains(c.EffectiveColors(), ctx.TextColor("W")) {
				toughness = -4
			}
			return BoostUntilEOT{
				Target:    id,
				Power:     -4,
				Toughness: toughness,
				Label:     "Ray of Enfeeblement",
			}.Apply(ctx)
		},
	})
}
