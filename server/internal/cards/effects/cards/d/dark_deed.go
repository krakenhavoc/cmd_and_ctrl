package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dark Deed — {1}{B} Instant (Marvel Super Heroes, #4455):
//
//	"Target creature gets -4/-4 until end of turn."
//
// One target; the shrink is a BoostUntilEOT with negative values, so
// toughness 0 or less puts the creature in the graveyard as a state-based
// action (CR 704.5f). A target that left by resolution is skipped
// (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e98cd00c-b812-4c6a-81d4-e3f94b64fc58",
		Name:         "Dark Deed",
		Purpose:      effects.ForTargets(effects.HarmUntilEOT(0, -4, -4)),
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			return effects.BoostUntilEOT{
				Target:    item.Targets[0].ID,
				Power:     -4,
				Toughness: -4,
				Label:     "Dark Deed — -4/-4 until end of turn",
			}.Apply(ctx)
		},
	})
}
