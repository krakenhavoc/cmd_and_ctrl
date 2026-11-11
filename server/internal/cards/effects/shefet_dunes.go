package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(cmdr2B13DesertLand("8305715e-f711-47d6-8efe-d0efe4ced418", "Shefet Dunes", "W", ActivatedAbility{
		Label:        "{2}{W}{W}, {T}, Sacrifice a Desert: Creatures you control get +1/+1 until end of turn. Activate only as a sorcery.",
		Purpose:      game.Purpose{TeamPump: &game.Pump{Power: 1, Toughness: 1}},
		Cost:         Plus(ManaCost("{2}{W}{W}"), TapCost(), cmdr2B13SacrificeADesert()),
		SorcerySpeed: true,
		Effect: func(g *game.Game, item *game.StackItem) error {
			return BoostUntilEOT{
				Match: And(Creature(), YouControl()), Power: 1, Toughness: 1,
				Label: "Shefet Dunes — creatures you control get +1/+1",
			}.Apply(NewContext(g, item))
		},
	}))
}
