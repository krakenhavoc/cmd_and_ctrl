package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sadistic Slash — Instant {3}{B}:
//
//	"Target creature gets -5/-5 until end of turn.
//	 Mayhem {1}{B} (You may cast this card from your graveyard for
//	 {1}{B} if you discarded it this turn.)"
//
// Dismember's body with mayhem (CR 702.187, #4628). No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "0e471b6d-4ba4-46aa-a802-81e401cbfa56",
		Name:             "Sadistic Slash",
		Purpose:          effects.ForTargets(effects.HarmUntilEOT(0, -5, -5)),
		Completeness:     effects.CompletenessFull,
		Targets:          effects.TargetCreature("target creature"),
		CastableZonesIf:  []game.ConditionalCastZone{effects.MayhemCastableZone()},
		AlternativeCosts: []game.AlternativeCost{effects.Mayhem("{1}{B}")},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard || !ctx.IsTargetLegal(item.Targets[0]) {
				return nil
			}
			return effects.BoostUntilEOT{
				Target:    item.Targets[0].ID,
				Power:     -5,
				Toughness: -5,
				Label:     "Sadistic Slash — -5/-5",
			}.Apply(ctx)
		},
	})
}
