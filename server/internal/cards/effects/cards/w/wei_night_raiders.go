package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Wei Night Raiders — Creature — Human Soldier {2}{B}{B}, 2/2
// (Portal Three Kingdoms, #5622):
//
//	"Horsemanship (This creature can't be blocked except by creatures with horsemanship.)"
//	"Whenever this creature deals damage to an opponent, that player discards a card."
//
// Horsemanship is an engine keyword carried as printed data: the block
// check refuses any blocker without horsemanship. It doesn't stop this
// creature blocking.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "d24ee48a-86a9-4a72-9311-79b6ea3f41f2",
		Name:            "Wei Night Raiders",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"horsemanship"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDealsDamageToAnOpponent("Wei Night Raiders — that player discards a card", effects.ConB2ThatPlayerDiscards(1, "Wei Night Raiders — discard a card")),
		},
	})
}
