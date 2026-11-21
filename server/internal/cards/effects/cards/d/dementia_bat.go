package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dementia Bat — Creature — Phyrexian Bat {4}{B}, 2/2 (New Phyrexia, #5575):
//
//	"Flying
//	 {4}{B}, Sacrifice this creature: Target player discards two cards."
//
// Flying, and a sacrifice row whose target player is asked to discard
// two (targetOpponentDiscards reads any target player). A player with
// fewer cards discards what they have (CR 701.9b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "95c76175-e3a6-4bf4-a8da-d1179851efe5",
		Name:            "Dementia Bat",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{4}{B}, Sacrifice this creature: Target player discards two cards.",
			Cost:    effects.Plus(effects.ManaCost("{4}{B}"), effects.SacrificeThis()),
			Targets: effects.TargetPlayer("target player"),
			Effect:  effects.TargetOpponentDiscards(2),
		}},
	})
}
