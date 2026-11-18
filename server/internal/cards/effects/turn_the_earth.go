package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Turn the Earth — Instant {G} (Innistrad: Midnight Hunt, #5497):
//
//	"Choose up to three target cards in graveyards. The owners of those
//	 cards shuffle them into their libraries. You gain 2 life.
//	 Flashback {1}{G}"
//
// Those Old Scientists' chapter III with three targets: the cards may be
// in different graveyards, each goes to its OWNER's library in one move
// (a commander's owner is asked, CR 903.9) and each of those owners
// shuffles (CR 701.24c). The targets are "up to", so none may be chosen,
// and the spell still resolves and gains the life; once targets are
// chosen and every one is illegal on resolution, the spell does nothing
// at all (CR 608.2b). The life is gained by the caster.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "ebb445f7-75da-4af4-9153-93d419c838e0",
		Name:             "Turn the Earth",
		Completeness:     CompletenessFull,
		Targets:          TargetCardInGraveyard("up to three target cards in graveyards").WithCount(0, 3),
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Flashback("{1}{G}")},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if err := trekB9ShuffleTargetsIntoOwnersLibraries(ctx.Game, item); err != nil {
				return err
			}
			return GainLife{Amount: 2}.Apply(ctx)
		},
	})
}
