package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hulkbuster Armor — Legendary Artifact — Equipment {4} (Marvel Super Heroes
// Commander):
//
//	"Equipped creature has base power and toughness 9/9 and has flying.
//	 Equip Hero {3}
//	 Equip {6}"
//
// Blade of the Oni's pair: a layer-7b base-P/T set (SetAttachedBasePT, so
// counters and pumps still add on top) and a layer-6 keyword grant. The two
// equip abilities are Bilbo's Ring's shape, the restricted one narrowed to a
// Hero you control and the plain one to any creature you control.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "abc3f55e-5fc6-44b3-a2f5-b8f49d93249f",
		Name:         "Hulkbuster Armor",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.SetAttachedBasePT(9, 9),
			effects.EquippedCreatureIs(game.Layer6Ability, "Equipped creature has flying.", func(c *game.Characteristic) {
				c.Abilities = game.AppendKeywordAbility(c.Abilities, "flying")
			}),
		},
		Activated: []effects.ActivatedAbility{
			effects.EquipOnlyAbility("Equip Hero {3}", "{3}",
				effects.TargetCreature("target Hero you control", effects.YouControl(), effects.Subtype("Hero"))),
			effects.EquipAbility("{6}"),
		},
	})
}
