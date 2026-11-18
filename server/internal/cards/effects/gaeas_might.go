package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Gaea's Might — Instant {G}:
//
//	"Domain — Target creature gets +1/+1 until end of turn for each basic
//	 land type among lands you control."
//
// The count is read as the spell resolves (CR 611.2a), from the shared
// domain count (DomainCount, #5666).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "73b26f12-78eb-4d01-9dd6-ee643c7a80a8",
		Name:         "Gaea's Might",
		Purpose:      ForTargets(game.TargetPurpose{Slot: 0, UntilEndOfTurn: game.TargetBenefit}),
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			n := DomainCount(ctx.Game, item.Controller)
			return BoostUntilEOT{
				Target:    item.Targets[0].ID,
				Power:     n,
				Toughness: n,
				Label:     "Gaea's Might — +1/+1 for each basic land type",
			}.Apply(ctx)
		},
	})
}
