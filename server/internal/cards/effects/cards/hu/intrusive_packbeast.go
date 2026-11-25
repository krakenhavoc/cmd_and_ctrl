package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Intrusive Packbeast — Creature — Beast {4}{W}, 3/3 (Guilds of Ravnica,
// #5525):
//
//	"Vigilance
//	 When this creature enters, tap up to two target creatures your opponents
//	 control."
//
// Eddymurk Crab's tap of up to two targets, narrowed to creatures an opponent
// controls: zero, one or two picks, each re-checked as the trigger resolves
// (CR 608.2b).
//
// No simplification.
func init() {
	enters := effects.WhenThisEnters("Intrusive Packbeast — tap up to two target creatures your opponents control", effects.TapEachLegalTarget)
	enters.Targets = effects.TargetCreature("up to two target creatures your opponents control", effects.OpponentControls()).WithCount(0, 2)
	effects.Register(effects.Spec{
		OracleID:        "85a59833-b661-4599-a5d1-042dee033617",
		Name:            "Intrusive Packbeast",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered:       []game.TriggeredAbility{enters},
	})
}
