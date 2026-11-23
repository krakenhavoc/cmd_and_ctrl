package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Klingon Strike Force — Creature — Klingon Warrior {3}{R}{R}, 5/5 (Star
// Trek Commander, #4974):
//
//	"Trample
//	 Haste
//	 When this creature dies, it deals 2 damage to each opponent."
//
// Trample and haste are printed keywords. The dies trigger is one event
// of 2 damage from the Strike Force, as it last existed on the
// battlefield (CR 608.2h), to each opponent of the player who controlled
// it as it died.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "847efb61-5933-43c1-93ff-e8375f9469a9",
		Name:            "Klingon Strike Force",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample", "haste"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Klingon Strike Force — 2 damage to each opponent", effects.Do(effects.DamageEachOpponent{N: 2})),
		},
	})
}
