package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Cracked Earth Technique — Sorcery — Lesson {4}{G} (Avatar: The Last
// Airbender Eternal, #5075):
//
//	"Earthbend 3, then earthbend 3. You gain 3 life."
//
// Dai Li Agents' two earthbend clauses (CR 701.66a), each its own
// "target land you control"; one land may answer both (CR 115.3, the
// ruling), and then it gets six counters. The second earthbend waits for
// the first one's counters to land (a CR 616 prompt included), and the
// life gain waits for the second, each through a keyed tail. A land that
// has become illegal is skipped and the rest still happens; with both
// illegal the spell does not resolve (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d5314006-afe0-4cdc-88f0-ff1f7d85fba9",
		Name:         "Cracked Earth Technique",
		Completeness: CompletenessFull,
		Targets:      Clauses(EarthbendTargets(), EarthbendTargets()),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			first := crackedEarthLand(ctx, 0)
			if first == uuid.Nil {
				return crackedEarthSecond(ctx)
			}
			return Earthbend{Target: first, N: 3, Then: ctx.ItemTail(crackedEarthThenSecond)}.Apply(ctx)
		},
	})
}

var crackedEarthThenSecond = CardTail("avatar-b30/cracked-earth-second-earthbend", func(ctx *Context, _ game.TailOutcome) error {
	return crackedEarthSecond(ctx)
})

var crackedEarthThenLife = CardTail("avatar-b30/cracked-earth-gain-life", func(ctx *Context, _ game.TailOutcome) error {
	return GainLife{Amount: 3}.Apply(ctx)
})

// crackedEarthSecond is "then earthbend 3. You gain 3 life."
func crackedEarthSecond(ctx *Context) error {
	second := crackedEarthLand(ctx, 1)
	if second == uuid.Nil {
		return GainLife{Amount: 3}.Apply(ctx)
	}
	return Earthbend{Target: second, N: 3, Then: ctx.ItemTail(crackedEarthThenLife)}.Apply(ctx)
}

// crackedEarthLand is clause i's land while it is still a legal target
// on the battlefield, or uuid.Nil.
func crackedEarthLand(ctx *Context, i int) uuid.UUID {
	t, ok := ctx.ClauseTarget(i)
	if !ok || t.Kind != game.TargetCard || t.ID == uuid.Nil || !onBattlefield(ctx.Game, t.ID) {
		return uuid.Nil
	}
	return t.ID
}
