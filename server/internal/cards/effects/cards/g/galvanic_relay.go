package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Galvanic Relay — Sorcery {2}{R} (Modern Horizons 2, #5501):
//
//	"Exile the top card of your library. During your next turn, you may
//	 play that card.
//	 Storm (When you cast this spell, copy it for each spell cast before it
//	 this turn.)"
//
// The permission is open only during the holder's next turn (the grant's
// CondDuringYourNextTurn start, with the end-of-next-turn window as its
// end), so a card exiled on your own turn can't be played that turn. Every
// storm copy exiles its own card.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "973c684c-a396-4e40-8aaa-365bc64a6f9a",
		Name:         "Galvanic Relay",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{effects.Storm()},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			you := ctx.Controller()
			window := ctx.Game.UntilEndOfYourNextTurnDuration(you)
			return effects.ExileTopWithPermission{
				From: you, GrantTo: you, N: 1,
				Window:    &window,
				Condition: ctx.Game.YourNextTurnConditionForEffect(you),
			}.Apply(ctx)
		},
	})
}
