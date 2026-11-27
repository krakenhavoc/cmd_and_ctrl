package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Magebane Armor — Artifact — Equipment {3} (Magic 2010, #5585):
//
//	"Equipped creature gets +2/+4 and loses flying.
//	 Prevent all noncombat damage that would be dealt to equipped creature.
//	 Equip {2}"
//
// Starforged Sword's pump and loss of flying, Multiclass Baldric's prevention
// static narrowed to noncombat damage (CR 615.1).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fc24af7b-f3d0-41d0-aaac-d83424455717",
		Name:         "Magebane Armor",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.PumpAttached(2, 4),
			effects.RemoveFromAttached("flying"),
		},
		Replacements: []game.ReplacementEffect{
			effects.PreventAllDamageDealtTo(effects.PreventionStatic{
				To:     effects.ToEquippedCreature,
				Damage: effects.NoncombatDamage,
				Label:  "Magebane Armor: prevent all noncombat damage that would be dealt to equipped creature",
			}),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{2}")},
	})
}
