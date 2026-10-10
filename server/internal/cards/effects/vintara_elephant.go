package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Vintara Elephant — Creature — Elephant {4}{G}, 4/3:
//
//	"Trample
//	 {3}: This creature loses trample until end of turn. Any player may
//	 activate this ability."
//
// Trample is a printed keyword (CR 702.19). The activation is an
// any-player row (CR 602.2, 602.1b): whoever activates it pays the {3}
// out of their own pool (CR 602.1a), and the Elephant loses trample
// until end of turn. The loss is a layer-6 ability-removing effect with
// its own timestamp (CR 613.1f), so a trample granted LATER in the turn
// puts it back (effects_this_until_end_of_turn.go).
//
// No Purpose: the bot never pays to change a creature it does not
// control (ADR 0106 owner decision 2).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "0a73b4da-9b5c-480e-b58b-6a82a52ea71e",
		Name:            "Vintara Elephant",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Activated: []ActivatedAbility{{
			Label:     "{3}: This creature loses trample until end of turn. Any player may activate this ability.",
			Purpose:   game.Purpose{Answers: game.AnswerCombatGrant},
			Cost:      ManaCost("{3}"),
			AnyPlayer: true,
			Effect:    thisLosesKeywordUntilEndOfTurn("trample", "Vintara Elephant — loses trample until end of turn"),
		}},
	})
}
