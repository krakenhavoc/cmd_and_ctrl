package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Feeling of Dread — Instant {1}{W} (Innistrad, #5572):
//
//	"Tap up to two target creatures.
//	 Flashback {1}{U} (You may cast this card from your graveyard for
//	 its flashback cost. Then exile it.)"
//
// Repel the Darkness's tap (tapEachLegalTarget over up to two creatures)
// without the draw, with the shared Flashback ({1}{U}).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "269208e1-353c-4334-941a-2177f6665fba",
		Name:             "Feeling of Dread",
		Completeness:     CompletenessFull,
		Targets:          TargetCreature("up to two target creatures").WithCount(0, 2),
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Flashback("{1}{U}")},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return tapEachLegalTarget(ctx.Game, item)
		},
	})
}
