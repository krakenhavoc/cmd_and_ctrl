package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Keldon Warcaller — Creature — Human Warrior {1}{R}, 2/2 (Dominaria):
//
//	"Whenever this creature attacks, put a lore counter on target Saga you
//	 control."
//
// An attack trigger whose target, a Saga you control, is chosen as it goes on
// the stack (CR 603.3d) and checked again as it resolves (CR 608.2b). The lore
// counter is an ordinary counter placement, so the Saga's next chapter triggers
// as the counter lands (CR 714.2c) and a counter-doubling effect applies.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b44d3fe8-b1e2-4abb-ab82-7a516a1db1e5",
		Name:         "Keldon Warcaller",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WheneverThisAttacks("Keldon Warcaller — put a lore counter on target Saga you control", effects.OldB209LoreCounterOnTarget),
				effects.TargetPermanent("target Saga you control", effects.HasSubtype("Saga"), effects.YouControl())),
		},
	})
}
