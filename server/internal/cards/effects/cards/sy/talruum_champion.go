package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Talruum Champion — Creature — Minotaur {4}{R}, 3/3 (Visions, S82 older sets
// batch 791, #5631):
//
//	"First strike
//	 Whenever this creature blocks or becomes blocked by a creature, that
//	 creature loses first strike until end of turn."
//
// A trigger for each blocker-attacker pair the Champion is in (Mirror Shield's
// shape, CR 509.3b, 509.3d): when it blocks, the attacker it blocks; when it
// becomes blocked, each creature that blocks it. The creature is named, with its
// object epoch, as the trigger goes on the stack and loses first strike
// (a layer-6 removal with its own timestamp) only while it is still that
// creature on the battlefield (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "9352e742-9a22-4463-b41f-3117ff1540ed",
		Name:            "Talruum Champion",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Triggered: []game.TriggeredAbility{
			effects.Old791TalruumChampionPairs("Talruum Champion — that creature loses first strike until end of turn"),
		},
	})
}
