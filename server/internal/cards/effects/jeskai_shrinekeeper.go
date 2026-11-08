package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Jeskai Shrinekeeper — Creature — Dragon {2}{U}{R}{W}, 3/3 (Tarkir:
// Dragonstorm, #4463):
//
//	"Flying, haste
//	 Whenever this creature deals combat damage to a player, you gain 1
//	 life and draw a card."
//
// The combat-damage trigger is WheneverThisDealsCombatDamageToAPlayer, and
// its body Inspiring Overseer's: the life, then the draw, which waits for
// a prompt the life gain asks (CR 608.2c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "bb50b4f4-f0da-4283-821b-6d74bd20752f",
		Name:            "Jeskai Shrinekeeper",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "haste"},
		Triggered: []game.TriggeredAbility{
			WheneverThisDealsCombatDamageToAPlayer("Jeskai Shrinekeeper — you gain 1 life and draw a card",
				Do(GainLife{Amount: 1}, DrawCards{N: 1})),
		},
	})
}
