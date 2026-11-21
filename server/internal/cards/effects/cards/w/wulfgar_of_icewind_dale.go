package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wulfgar of Icewind Dale — Legendary Creature — Human Barbarian {3}{R}{G},
// 4/4 (Adventures in the Forgotten Realms):
//
//	"Melee (Whenever this creature attacks, it gets +1/+1 until end of turn
//	 for each opponent you attacked this combat.)
//	 If a creature you control attacking causes a triggered ability of a
//	 permanent you control to trigger, that ability triggers an additional
//	 time."
//
// Melee is the engine keyword. The second line is Isshin's attacking
// doubler narrowed to creatures its controller controls (CR 603.2d), so it
// doubles Wulfgar's own melee and every attack watcher on a permanent you
// control, but not an opponent's attacker. No simplification.
func init() {
	doubler := effects.DoublesAttacking(effects.And(effects.Creature(), effects.YouControl()))
	doubler.Label = "Wulfgar of Icewind Dale"
	doubler.RowLabel = "Abilities of your permanents that trigger when a creature you control attacks trigger an additional time."
	effects.Register(effects.Spec{
		OracleID:        "b19d1459-eb08-4e75-95f4-402d86371a78",
		Name:            "Wulfgar of Icewind Dale",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordMelee},
		TriggerDoublers: []game.TriggerDoubler{doubler},
	})
}
