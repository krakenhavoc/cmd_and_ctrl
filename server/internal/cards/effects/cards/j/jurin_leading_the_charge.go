package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Jurin, Leading the Charge — Legendary Creature — Human Barbarian {4}{R}{R},
// 4/6 (Secret Lair Drop):
//
//	"Haste
//	 Jurin must be blocked if able.
//	 Whenever Jurin attacks, each creature you control attacking a player gets
//	 +1/+0 until end of turn for each creature that player controls."
//
// Haste is printed data; the block requirement is the shared must-be-blocked
// static. The trigger pumps each attacker by the creatures of the player it
// attacks, counted as it resolves; an attacker aimed at a planeswalker or battle
// is not attacking a player and gets nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "de384867-8bb4-4edf-b542-970538718fdd",
		Name:            "Jurin, Leading the Charge",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Static:          []game.StaticAbility{effects.MustBeBlocked()},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Jurin, Leading the Charge — each creature you control attacking a player gets +1/+0 for each creature that player controls", effects.S82OldB024JurinPump),
		},
	})
}
