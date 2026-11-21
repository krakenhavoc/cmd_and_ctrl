package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tribal Unity — Instant {X}{2}{G}:
//
//	"Creatures of the creature type of your choice get +X/+X until end
//	 of turn."
//
// Chosen as the spell resolves (#2382). It pumps EVERY creature of the
// type, yours and everyone else's, and the set is fixed when the answer
// arrives (CR 611.2c): a creature that enters later this turn gets
// nothing. A changeling is of every type and is always in the set.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6b4ee935-4474-4757-9fab-978a2e1e72a0",
		Name:         "Tribal Unity",
		XMatters:     true,
		Completeness: effects.CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			effects.ChooseCreatureTypeTail(ctx.Game, item.Controller, item.SourceCardID, "Tribal Unity — choose a creature type",
				effects.WithN(ctx.ItemTail(tribalUnityTypeChosen), ctx.X()))
			return nil
		},
	})
}

// tribalUnityTypeChosen gives creatures of the chosen type +X/+X until end
// of turn, X the tail's N. An on-disk key.
var tribalUnityTypeChosen = effects.CreatureTypeChosen("then/tribal-unity-type-chosen", func(ctx *effects.Context, t string, o game.TailOutcome) error {
	return effects.BoostUntilEOT{
		Match:     effects.OfCreatureType(t),
		Power:     o.Tail.N,
		Toughness: o.Tail.N,
		Label:     "Tribal Unity — +X/+X",
	}.Apply(ctx)
})
