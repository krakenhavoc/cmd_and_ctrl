package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cubwarden — Creature — Cat {3}{W}, 3/5 (Ikoria: Lair of Behemoths, #5512):
//
//	"Mutate {2}{W}{W} (If you cast this spell for its mutate cost, put it over
//	 or under target non-Human creature you own. They mutate into the creature
//	 on top plus all abilities from under it.)
//	 Lifelink
//	 Whenever this creature mutates, create two 1/1 white Cat creature tokens
//	 with lifelink."
//
// Mutate is the engine's (ADR 0223); the trigger fires on the merge whichever
// card ended on top (CR 702.140d), and the merged creature has lifelink from
// Cubwarden's card wherever it sits (CR 702.140e).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "8c1b3f1b-40d8-4cd5-bedf-92c3c534be84",
		Name:             "Cubwarden",
		Completeness:     effects.CompletenessFull,
		PrintedKeywords:  []string{"lifelink"},
		Purpose:          game.Purpose{Tokens: 2, CreatureTokens: 2},
		AlternativeCosts: []game.AlternativeCost{effects.Mutate("{2}{W}{W}")},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventMutate, effects.ThisCreatureMutates, "Cubwarden — create two 1/1 white Cat creature tokens with lifelink",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 white Cat with lifelink"), N: 2})),
		},
	})
}
