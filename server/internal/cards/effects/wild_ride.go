package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wild Ride — Sorcery {R} (#3931):
//
//	"Target creature gets +3/+0 and gains haste until end of turn.
//	 Harmonize {4}{R}"
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "61978202-538b-4f92-acfc-e95f4a6ce466",
		Name:             "Wild Ride",
		Completeness:     CompletenessFull,
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Harmonize("{4}{R}")},
		Targets:          TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return firstTargetGainsUntilEndOfTurn(3, "Wild Ride — +3/+0 and haste", "haste")(ctx.Game, item)
		},
	})
}
