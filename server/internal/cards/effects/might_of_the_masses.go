package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Might of the Masses — Instant {G}:
//
//	"Target creature gets +1/+1 until end of turn for each creature you
//	 control."
//
// The creatures are counted as the spell resolves, so the target itself
// counts when you control it and a creature that died in response does
// not. A target that is illegal means the spell does nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f73416a8-40e5-4791-948f-147ef2221ee7",
		Name:         "Might of the Masses",
		Completeness: CompletenessFull,
		Purpose:      ForTargets(BenefitUntilEOT(0, 0, 0)),
		Targets:      TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			legal := ctx.LegalTargets()
			if len(legal) == 0 || legal[0].Kind != game.TargetCard {
				return nil
			}
			n := len(permanentsControlledByMatching(ctx.Game, item.Controller, Creature()))
			return BoostUntilEOT{Target: legal[0].ID, Power: n, Toughness: n, Label: "Might of the Masses — +1/+1 per creature"}.Apply(ctx)
		},
	})
}
