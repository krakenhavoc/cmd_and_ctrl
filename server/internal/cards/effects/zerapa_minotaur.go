package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Zerapa Minotaur — Creature — Minotaur {2}{R}{R}, 3/3:
//
//	"First strike
//	 {2}: This creature loses first strike until end of turn. Any
//	 player may activate this ability."
//
// First strike is a printed keyword (CR 702.7). The activation is an
// any-player row (CR 602.2, 602.1b): whoever activates it pays the {2}
// out of their own pool (CR 602.1a), and the Minotaur loses first
// strike until end of turn. The loss is a layer-6 ability-removing
// effect with its own timestamp (CR 613.1f), so a first strike granted
// LATER in the turn puts it back and one granted earlier does not
// (effects_this_until_end_of_turn.go).
//
// No Purpose: the bot never pays to change a creature it does not
// control (ADR 0106 owner decision 2).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "4520c630-3650-462f-ab01-215eb2bdbbf5",
		Name:            "Zerapa Minotaur",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Activated: []ActivatedAbility{{
			Label:     "{2}: This creature loses first strike until end of turn. Any player may activate this ability.",
			Purpose:   game.Purpose{Answers: game.AnswerCombatGrant},
			Cost:      ManaCost("{2}"),
			AnyPlayer: true,
			Effect:    thisLosesKeywordUntilEndOfTurn("first strike", "Zerapa Minotaur — loses first strike until end of turn"),
		}},
	})
}
