package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Trumpeting Gnarr — Creature — Beast {1}{G}{U}, 3/3 (Ikoria: Lair of
// Behemoths, #5512):
//
//	"Mutate {3}{G/U}{G/U} (If you cast this spell for its mutate cost, put it
//	 over or under target non-Human creature you own. They mutate into the
//	 creature on top plus all abilities from under it.)
//	 Whenever this creature mutates, create a 3/3 green Beast creature token."
//
// Mutate is the engine's (ADR 0223); the trigger fires on the merge whichever
// card ended on top (CR 702.140d).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "734be5ec-6148-48b9-9021-380e9b62ec23",
		Name:             "Trumpeting Gnarr",
		Completeness:     effects.CompletenessFull,
		Purpose:          game.Purpose{Tokens: 1, CreatureTokens: 1},
		AlternativeCosts: []game.AlternativeCost{effects.Mutate("{3}{G/U}{G/U}")},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventMutate, effects.ThisCreatureMutates, "Trumpeting Gnarr — create a 3/3 green Beast creature token",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("3/3 green Beast"), N: 1})),
		},
	})
}
