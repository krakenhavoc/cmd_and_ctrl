package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kalemne's Captain — Creature — Giant Soldier {3}{W}{W}, 5/5 (Commander
// 2015, #5546):
//
//	"Vigilance
//	 {5}{W}{W}: Monstrosity 3. (If this creature isn't monstrous, put three
//	 +1/+1 counters on it and it becomes monstrous.)
//	 When this creature becomes monstrous, exile all artifacts and
//	 enchantments."
//
// Monstrosity is the shared activated ability (Arbor Colossus' shape) and
// the trigger WhenBecomesMonstrous (CR 701.37b). The exile is Farewell's
// ExileAllMatching over every artifact and enchantment on the battlefield,
// its controller's included.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "663e6ab0-ed99-4cec-86f7-843207c31e17",
		Name:            "Kalemne's Captain",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Purpose:         game.Purpose{Sweep: game.Sweep{Matches: game.SweepArtifactsAndEnchantments, How: game.SweepExile}},
		Activated:       []effects.ActivatedAbility{effects.Monstrosity(effects.ManaCost("{5}{W}{W}"), 3)},
		Triggered: []game.TriggeredAbility{
			effects.WhenBecomesMonstrous("Kalemne's Captain — exile all artifacts and enchantments",
				effects.Do(effects.ExileAllMatching{Match: effects.Or(effects.Artifact(), effects.Enchantment())})),
		},
	})
}
