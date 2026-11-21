package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Moonsilver Spear — {4} Artifact — Equipment:
//
//	"Equipped creature has first strike.
//	 Whenever equipped creature attacks, create a 4/4 white Angel
//	 creature token with flying.
//	 Equip {4}"
//
// Glamdring's first-strike grant and Nuka-Nuke Launcher's attack trigger
// on the equipped creature. The Angel is created by the Spear's
// controller, who need not control the equipped creature's attack.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d99bcf62-db93-40db-8844-3d0af0756a0c",
		Name:         "Moonsilver Spear",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.GrantToAttached("first strike"),
		},
		Triggered: []game.TriggeredAbility{
			effects.WheneverEnchantedCreatureAttacks("Moonsilver Spear — create a 4/4 white Angel creature token with flying",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("4/4 white Angel with flying"), N: 1})),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{4}")},
	})
}
