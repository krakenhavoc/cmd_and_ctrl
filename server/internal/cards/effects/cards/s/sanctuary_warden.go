package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sanctuary Warden — Creature — Angel Soldier {4}{W}{W}, 5/5 (Streets
// of New Capenna):
//
//	"Flying
//	 This creature enters with two shield counters on it. (If it would
//	 be dealt damage or destroyed, remove a shield counter from it
//	 instead.)
//	 Whenever this creature enters or attacks, you may remove a counter
//	 from a creature or planeswalker you control. If you do, draw a
//	 card and create a 1/1 green and white Citizen creature token."
//
// One ability, two conditions (WhenThisEntersOrAttacks). As it resolves
// the controller says no, or names the permanent and the kind of
// counter in one question; the Warden's own counters are on the list.
// Nothing to remove means nothing is drawn (CR 608.2c "if you do").
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "713bcfb6-e5ea-411e-b4b3-52e5eed5cdf1",
		Name:            "Sanctuary Warden",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Purpose:         game.Purpose{Tokens: 1, CreatureTokens: 1},
		Replacements:    []game.ReplacementEffect{effects.Snc6EntersWithShield("Sanctuary Warden", 2)},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEntersOrAttacks("Sanctuary Warden — you may remove a counter from a creature or planeswalker you control; if you do, draw a card and create a Citizen",
				effects.Do(effects.Instruction(effects.Snc6WardenRemove))),
		},
	})
}
