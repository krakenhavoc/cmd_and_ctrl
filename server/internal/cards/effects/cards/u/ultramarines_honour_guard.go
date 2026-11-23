package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ultramarines Honour Guard — Creature — Astartes Warrior {3}{W}, 2/2
// (Warhammer 40,000 Commander, S82 GAP batch 196, #5948):
//
//	"Squad {2} (As an additional cost to cast this spell, you may pay {2} any
//	 number of times. When this creature enters, create that many tokens that are
//	 copies of it.)
//	 Other creatures you control get +1/+1."
//
// Squad is Wasteland Raider's optional cost and enters trigger. The anthem is
// Lumen-Class Frigate's: a layer 7c +1/+1 on every other creature its controller
// controls, so each squad copy pumps the rest, the other copies included.
//
// No simplification.
func init() {
	anthem := effects.TribalAnthem(effects.TribeFilter{Others: true, YoursOnly: true}, 1, 1)
	anthem.Label = "Other creatures you control get +1/+1."
	effects.Register(effects.Spec{
		OracleID:      "74cd099c-8476-4f75-bc1a-8b166f696163",
		Name:          "Ultramarines Honour Guard",
		Completeness:  effects.CompletenessFull,
		OptionalCosts: []game.AdditionalCost{effects.SquadMana("{2}", 20)},
		Triggered:     []game.TriggeredAbility{effects.Squad("Ultramarines Honour Guard")},
		Static:        []game.StaticAbility{anthem},
	})
}
