package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Lawless Broker — Creature — Aetherborn Rogue {2}{B}, 3/2 (Jumpstart,
// #5510):
//
//	"When this creature dies, put a +1/+1 counter on target creature you
//	 control."
//
// Bat Colony's counter body on a dies trigger; the Broker is already in the
// graveyard when the target is chosen, so it can't pick itself.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d17d085f-aeda-4b23-ab80-729d070ae813",
		Name:         "Lawless Broker",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisDies("Lawless Broker — put a +1/+1 counter on target creature you control", effects.Cmdr2S46PlusOneCounterOnClauseTarget),
				effects.TargetCreature("target creature you control", effects.YouControl())),
		},
	})
}
