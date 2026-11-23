package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cruel Administrator — Creature — Human Soldier {3}{B}{R}, 5/4 (Avatar:
// The Last Airbender, #4459):
//
//	"Raid — This creature enters with a +1/+1 counter on it if you
//	 attacked this turn.
//	 Whenever this creature attacks, create a 2/2 red Soldier creature
//	 token with firebending 1. (Whenever it attacks, add {R}. This mana
//	 lasts until end of combat.)"
//
// Raid is a self-replacement on the entry (CR 614.1c), read as the
// creature enters: "you attacked this turn" is the turn's attack tally
// (b18AttackedThisTurn). The Soldier is the printed ttla token, its
// firebending a triggered ability of its own (ADR 0083); it is not
// attacking, so it does not firebend this combat.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "24a691f3-88dc-47a5-b5cd-e02bb8026470",
		Name:         "Cruel Administrator",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.AvB16SelfEntersWithCountersIfRaid(game.CounterPlusOne, 1)},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Cruel Administrator — create a 2/2 red Soldier with firebending 1",
				effects.Do(effects.CreateToken{Template: effects.AvB16SoldierFirebendingToken(), N: 1})),
		},
	})
}
