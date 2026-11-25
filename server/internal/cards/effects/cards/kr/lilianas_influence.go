package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Liliana's Influence — Sorcery {4}{B}{B} (Amonkhet, S82 OLD batch 242, #5535):
//
//	"Put a -1/-1 counter on each creature you don't control. You may search
//	 your library and/or graveyard for a card named Liliana, Death Wielder,
//	 reveal it, and put it into your hand. If you search your library this way,
//	 shuffle."
//
// The counters go on first, to each creature an opponent controls; then the
// "you may" search, which asks for the zones before showing anything and
// shuffles only if the library was searched (Ajani's Aid's shared search).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7fe67cef-38c8-4071-86df-71053818d908",
		Name:         "Liliana's Influence",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepMinus, Amount: 1, OpponentsOnly: true}},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			for _, c := range effects.MatchingBattlefield(ctx, effects.And(effects.OpponentControls(), effects.Creature())) {
				if err := (effects.AddCounter{Target: c.InstanceID, Kind: game.CounterMinusOne, N: 1}).Apply(ctx); err != nil {
					return err
				}
			}
			return effects.MayChoice{
				Question: "Liliana's Influence — search your library and/or graveyard for a card named Liliana, Death Wielder?",
				Then:     effects.MaySearchNamedLibraryOrGraveyard,
				Word:     "Liliana, Death Wielder",
			}.Apply(ctx)
		},
	})
}
