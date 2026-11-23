package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dwarven Castle Guard — Creature — Dwarf Soldier {1}{W}, 2/1 (Final
// Fantasy, #4462):
//
//	"When this creature dies, create a 1/1 colorless Hero creature
//	 token."
//
// A leaves-the-battlefield trigger (CR 603.6c, 603.10a) that makes the
// Hero job select makes. It looks back in time, so it fires whatever
// killed the Guard, and the Hero comes for its controller as it died.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b51e0ae3-126b-4b7e-9dd1-818a5b8e512a",
		Name:         "Dwarven Castle Guard",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Dwarven Castle Guard — create a 1/1 colorless Hero creature token",
				effects.Do(effects.CreateToken{Template: effects.TokenCard(effects.JobSelectHeroToken), N: 1})),
		},
	})
}
