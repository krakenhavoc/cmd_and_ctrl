package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Monologue Tax — Enchantment {2}{W} (Marvel Super Heroes Commander,
// #4455):
//
//	"Whenever an opponent casts their second spell each turn, you create a
//	 Treasure token."
//
// Mangara, the Diplomat's second-spell read (mangaraOpponentCastTheirSecond):
// the cast tally is bumped before the cast event fires, so the trigger
// fires once, on the cast that takes that opponent to exactly two. Your
// own spells never count.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ace8e88e-ca6a-4bdf-8c6d-1b914fc00259",
		Name:         "Monologue Tax",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, mangaraOpponentCastTheirSecond,
				"Monologue Tax — an opponent cast their second spell: create a Treasure token",
				Do(CreateToken{Template: TreasureToken(), N: 1})),
		},
	})
}
