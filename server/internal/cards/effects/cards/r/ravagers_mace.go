package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ravager's Mace — Artifact — Equipment {1}{B}{R} (Zendikar Rising, #5508):
//
//	"When this Equipment enters, attach it to target creature you control.
//	 Equipped creature gets +1/+0 for each creature in your party and has
//	 menace. (Your party consists of up to one each of Cleric, Rogue,
//	 Warrior, and Wizard.)
//	 Equip {2}{B}{R}"
//
// The party (CR 700.8) is counted live, inside the layer pass, from the board
// as the earlier layers left it (Multiclass Baldric's read), so it follows the
// creatures coming and going and the equipped creature itself counts.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f3e0aaa1-f624-4dc7-a425-7dbf27cba1ce",
		Name:         "Ravager's Mace",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.PumpAttachedPer(1, 0, func(g *game.Game, source *game.Card) int {
				return effects.PartySizeInLayers(g, source.Controller)
			}),
			effects.GrantToAttached("menace"),
		},
		Triggered: []game.TriggeredAbility{effects.AfrB06EnterAttach("Ravager's Mace")},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{2}{B}{R}")},
	})
}
