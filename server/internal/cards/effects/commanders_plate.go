package effects

import (
	"slices"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Commander's Plate — Artifact — Equipment {1}:
//
//	"Equipped creature gets +3/+3 and has protection from each color
//	 that's not in your commander's color identity.
//	 Equip commander {3}
//	 Equip {5}"
//
// TWO EQUIP ABILITIES, the Blackblade Reforged shape: a cheaper one
// narrowed to the caster's own commander (YourCommander(), new for
// this card — see targets.go), and the plain {5} for any other
// creature. Both are ordinary Spec.Activated entries; nothing new
// there.
//
// The protection clause (#3172) is a layer 6 static reading the PLATE's
// controller's commander identity (CR 109.5: "your" is the Plate's
// controller, not the wearer's) through game.CommanderIdentityColorsForEffect
// and granting "protection from <colour>" for each of the five colours
// outside it. A controller with no commander has a colourless identity, so
// the creature is protected from every colour, as the rules read it
// (CR 903.4f); an identity the engine cannot read grants nothing, never
// more than printed.
func init() {
	Register(Spec{
		OracleID:     "cae166de-e681-40a0-83a8-3c17cf40e2fc",
		Name:         "Commander's Plate",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			PumpAttached(3, 3),
			{
				Label:     "Equipped creature has protection from each color that's not in your commander's color identity.",
				Layer:     game.Layer6Ability,
				AppliesTo: AttachedToSource,
				Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
					identity, known := g.CommanderIdentityColorsForEffect(source.Controller)
					if !known {
						return
					}
					for _, col := range []string{"W", "U", "B", "R", "G"} {
						if !slices.Contains(identity, col) {
							c.Abilities = game.AppendKeywordAbility(c.Abilities, game.ProtectionFromColor(col))
						}
					}
				},
			},
		},
		Activated: []ActivatedAbility{
			EquipOnlyAbility("Equip commander {3}", "{3}",
				TargetCreature("target creature you control that's your commander", YouControl(), YourCommander())),
			EquipAbility("{5}"),
		},
	})
}
