package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "d99bcf62-db93-40db-8844-3d0af0756a0c",
		Name:         "Moonsilver Spear",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			GrantToAttached("first strike"),
		},
		Triggered: []game.TriggeredAbility{
			WheneverEnchantedCreatureAttacks("Moonsilver Spear — create a 4/4 white Angel creature token with flying",
				Do(CreateToken{Template: TokenCard("4/4 white Angel with flying"), N: 1})),
		},
		Activated: []ActivatedAbility{EquipAbility("{4}")},
	})
}
