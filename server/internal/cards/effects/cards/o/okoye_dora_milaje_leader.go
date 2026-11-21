package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Okoye, Dora Milaje Leader — {3}{W} Legendary Creature — Human Warrior Hero
// 3/2 (Marvel Super Heroes, #4455):
//
//	"When Okoye enters, create two 1/1 white Soldier creature tokens.
//	 Attacking creature tokens you control have first strike."
//
// The enter trigger is the ordinary token-making ETB. The static is the
// "attacking creatures you control have <keyword>" layer 6 grant
// (AttackingCreaturesHave, Ohran Frostfang's shape) narrowed to tokens; it
// covers any token you control, Okoye's own Soldiers or not, for exactly as
// long as it is attacking.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6d397d26-7e52-47b2-9880-2c10e90cfba4",
		Name:         "Okoye, Dora Milaje Leader",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Okoye — create two 1/1 white Soldiers",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 white Soldier"), N: 2})),
		},
		Static: []game.StaticAbility{
			effects.AttackingCreaturesHave("Attacking creature tokens you control have first strike.",
				func(c *game.Card) bool { return c.IsToken() }, "first strike"),
		},
	})
}
