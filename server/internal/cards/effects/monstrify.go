package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Monstrify — Sorcery {3}{G}:
//
//	"Target creature gets +4/+4 until end of turn.
//	 Retrace"
//
// Retrace (CR 702.81) is a graveyard cast for the PRINTED mana cost plus
// a land card discarded from hand: `CastableZones` opens the graveyard and
// `Retrace(cost)` prices it (ADR 0066, 2026-10-07 amendment). The spell goes
// back to the graveyard when it resolves, so it can be retraced again.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "0f58f791-469a-4a22-996e-4906c0914858",
		Name:             "Monstrify",
		Completeness:     CompletenessFull,
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Retrace("{3}{G}")},
		Targets:          TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			return BoostUntilEOT{
				Target:    item.Targets[0].ID,
				Power:     4,
				Toughness: 4,
				Label:     "Monstrify — +4/+4",
			}.Apply(ctx)
		},
	})
}
