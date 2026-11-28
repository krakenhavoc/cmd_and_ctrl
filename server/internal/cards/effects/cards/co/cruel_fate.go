package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cruel Fate — Sorcery {4}{U} (Portal, S82 OLD batch 781, #5630):
//
//	"Look at the top five cards of target opponent's library. Put one of
//	 those cards into that player's graveyard and the rest on top of their
//	 library in any order."
//
// Wu Spy's look at another player's library (only the caster learns the cards)
// with a mandatory pick for the graveyard, then the rest back on top in an
// order the caster chooses (CR 401.4). A target opponent gone by resolution
// makes it do nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e199e183-e857-4dc7-87f4-36b32c4dac96",
		Name:         "Cruel Fate",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target opponent", effects.Opponent()),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.S82OldB781CruelFateLook(ctx.Game, item)
		},
	})
}
