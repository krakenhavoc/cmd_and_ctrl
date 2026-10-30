package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sadistic Slash — Instant {3}{B}:
//
//	"Target creature gets -5/-5 until end of turn.
//	 Mayhem {1}{B} (You may cast this card from your graveyard for
//	 {1}{B} if you discarded it this turn.)"
//
// Dismember's body with mayhem (CR 702.187, #4628). No simplification.
func init() {
	Register(Spec{
		OracleID:         "0e471b6d-4ba4-46aa-a802-81e401cbfa56",
		Name:             "Sadistic Slash",
		Purpose:          ForTargets(HarmUntilEOT(0, -5, -5)),
		Completeness:     CompletenessFull,
		Targets:          TargetCreature("target creature"),
		CastableZonesIf:  []game.ConditionalCastZone{MayhemCastableZone()},
		AlternativeCosts: []game.AlternativeCost{Mayhem("{1}{B}")},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard || !ctx.IsTargetLegal(item.Targets[0]) {
				return nil
			}
			return BoostUntilEOT{
				Target:    item.Targets[0].ID,
				Power:     -5,
				Toughness: -5,
				Label:     "Sadistic Slash — -5/-5",
			}.Apply(ctx)
		},
	})
}
