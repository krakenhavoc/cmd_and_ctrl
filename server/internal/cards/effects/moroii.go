package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Moroii — Creature — Vampire {2}{U}{B}, 4/4 (Ravnica: City of Guilds,
// #5600):
//
//	"Flying
//	 At the beginning of your upkeep, you lose 1 life."
//
// Blessing of Leeches' upkeep loss: a life change, not damage, at the
// beginning of its controller's upkeep.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "418be12f-b40a-4018-86c1-98c9000a999a",
		Name:            "Moroii",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep("Moroii — you lose 1 life", Do(GainLife{Amount: -1})),
		},
	})
}
