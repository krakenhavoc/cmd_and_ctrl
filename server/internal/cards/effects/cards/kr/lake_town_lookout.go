package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lake-town Lookout — Creature — Human Scout {W}, 1/1:
//
//	"When this creature dies, recruit. (Draw a card, then discard a card.
//	 If you discarded a nonland card, create a 1/1 white Human Soldier
//	 creature token.)"
//
// Recruit is CR 701.70a (Recruit, recruit.go), for the Lookout's
// controller as it died.
//
// No simplification. The Hobbit (#4475).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cf765efe-884c-48e2-9edb-9d45cf2756dd",
		Name:         "Lake-town Lookout",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Lake-town Lookout — recruit", effects.Do(effects.Recruit{})),
		},
	})
}
