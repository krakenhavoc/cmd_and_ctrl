package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hexgold Hoverwings — Artifact — Equipment {3}{W} (Phyrexia: All Will Be One):
//
//	"For Mirrodin! (When this Equipment enters, create a 2/2 red Rebel
//	 creature token, then attach this to it.)
//	 Equipped creature has flying.
//	 Creatures you control that are equipped get +1/+0.
//	 Equip {2}{W}"
//
// #5654, S82. The shared ForMirrodin trigger, a layer 6 flying grant and
// one equip ability. The second sentence is not about this Equipment's
// own creature: it is every creature its controller controls with any
// Equipment attached (isEquippedCreature, CR 301.5), so the Hoverwings'
// own host gets it only while its controller is also the Hoverwings'
// controller, and two Hoverwings give each equipped creature +2/+0.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "63018d1c-f7fb-44c8-bcd2-292e74ab5b54",
		Name:         "Hexgold Hoverwings",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{effects.ForMirrodin("Hexgold Hoverwings")},
		Static: []game.StaticAbility{
			effects.GrantToAttached("flying"),
			effects.Labelled(effects.B16Anthem(hexgoldHoverwingsEquippedCreatureYouControl, 1, 0),
				"Creatures you control that are equipped get +1/+0."),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{2}{W}")},
	})
}

// hexgoldHoverwingsEquippedCreatureYouControl is "creatures you control
// that are equipped": a creature the source's controller controls with an
// Equipment attached.
func hexgoldHoverwingsEquippedCreatureYouControl(target *game.Card, g *game.Game, source *game.Card) bool {
	return target.Controller == source.Controller && effects.IsEquippedCreature(g, *target)
}
