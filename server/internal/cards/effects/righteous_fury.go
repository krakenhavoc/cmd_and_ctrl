package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Righteous Fury — Sorcery {4}{W}{W}:
//
//	"Destroy all tapped creatures. You gain 2 life for each creature
//	 destroyed this way."
//
// Fumigate's sweep over the tapped creatures of every player, yours
// included, and its "destroyed this way" count (a creature that was
// indestructible, regenerated or was exiled instead is not counted;
// CR 701.8a), at 2 life each instead of 1. Tapped is read as the sweep
// resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f0ff2c1c-702d-406b-a426-4282117afe8e",
		Name:         "Righteous Fury",
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy}},
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return DestroyAllMatching{
				Match:    And(Creature(), tappedPermanent()),
				ThenTail: ctx.Tail(massGainTwoLifeForEach),
			}.Apply(ctx)
		},
	})
}

// massGainTwoLifeForEach is "you gain 2 life for each <permanent> <verb>
// this way" (Righteous Fury).
var massGainTwoLifeForEach = MassTail("mass/gain-two-life-for-each", func(ctx *Context, _ []game.Card, moved int) error {
	if moved <= 0 {
		return nil
	}
	return GainLife{Player: ctx.Controller(), Amount: 2 * moved}.Apply(ctx)
})
