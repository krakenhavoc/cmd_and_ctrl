package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// River Herald Scout — Creature — Merfolk Scout {1}{U}, 1/2 (The Lost
// Caverns of Ixalan, #4470):
//
//	"When this creature enters, it explores."
//
// Pathfinding Axejaw's trigger: the explorer is the object that entered
// (thatCreatureExplores). A land on top goes to the hand; anything else
// gives the Scout a +1/+1 counter and asks whether the card goes to the
// graveyard (CR 701.44a). A Scout that has left by resolution still
// reveals the card, and no counter is put on anything (CR 701.44c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "021428b1-005c-40ed-9d47-fb5aa970598b",
		Name:         "River Herald Scout",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("River Herald Scout — it explores", effects.ThatCreatureExplores),
		},
	})
}
