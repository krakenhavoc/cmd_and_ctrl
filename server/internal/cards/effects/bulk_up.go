package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bulk Up — Instant {1}{R}:
//
//	"Double target creature's power until end of turn.
//	 Flashback {4}{R}{R} (You may cast this card from your graveyard
//	 for its flashback cost. Then exile it.)"
//
// "Double" is a one-time snapshot, not a continuously recomputing
// multiplier (the Oracle ruling): this adds power equal to the
// target's CURRENT power (post-layer, counters included) as the
// spell resolves, through the same turn-scoped BoostUntilEOT registry
// Giant Growth uses. A creature that gets bigger or smaller later in
// the turn does not get doubled again — it is a flat +N/+0 for the
// N read right here.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "fade8af0-5fdb-4237-9dd6-48bfd1d62767",
		Name:             "Bulk Up",
		Completeness:     CompletenessFull,
		Targets:          TargetCreature("target creature"),
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Flashback("{4}{R}{R}")},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			id, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			c, ok := ctx.Game.LookupCardForEffect(id)
			if !ok {
				return nil
			}
			return BoostUntilEOT{
				Target: id,
				Power:  c.CurrentPower(),
				Label:  "Bulk Up — double power",
			}.Apply(ctx)
		},
	})
}
