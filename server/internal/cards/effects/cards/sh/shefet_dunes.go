package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shefet Dunes — Land — Desert:
//
//	"{T}: Add {C}.
//	 {T}, Pay 1 life: Add {W}.
//	 {2}{W}{W}, {T}, Sacrifice a Desert: Creatures you control get +1/+1
//	 until end of turn. Activate only as a sorcery."
//
// The pump affects the creatures you control as the ability resolves
// (CR 611.2c): a creature that arrives later is not pumped.
//
// No simplification.
func init() {
	effects.Register(effects.Cmdr2B13DesertLand("8305715e-f711-47d6-8efe-d0efe4ced418", "Shefet Dunes", "W", effects.ActivatedAbility{
		Label:        "{2}{W}{W}, {T}, Sacrifice a Desert: Creatures you control get +1/+1 until end of turn. Activate only as a sorcery.",
		Purpose:      game.Purpose{TeamPump: &game.Pump{Power: 1, Toughness: 1}},
		Cost:         effects.Plus(effects.ManaCost("{2}{W}{W}"), effects.TapCost(), effects.Cmdr2B13SacrificeADesert()),
		SorcerySpeed: true,
		Effect: func(g *game.Game, item *game.StackItem) error {
			return effects.BoostUntilEOT{
				Match: effects.And(effects.Creature(), effects.YouControl()), Power: 1, Toughness: 1,
				Label: "Shefet Dunes — creatures you control get +1/+1",
			}.Apply(effects.NewContext(g, item))
		},
	}))
}
