package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Assemble the Players — Enchantment {1}{W} (Murders at Karlov Manor,
// #4469):
//
//	"You may look at the top card of your library any time.
//	 Once each turn, you may cast a creature spell with power 2 or less
//	 from the top of your library."
//
// The look is LibraryTopVisible (game.LibraryTopOwner). The cast is a
// standing top-of-library permission over creature spells with power 2 or
// less (PermissionFilter.PowerCapped, the spell's power as it is cast, so a
// disguise card cast face down is a 2/2 whatever it prints), spent once a
// turn through the one-use-per-type budget on "creature" (the only type the
// filter opens). The permission is the card's own, so a second Assemble the
// Players has a budget of its own (ruling). The card keeps its normal timing:
// a creature spell is cast at sorcery speed, in any player's turn only if it
// has flash.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:          "180ba940-2b11-4953-b7ce-74847e1aebfd",
		Name:              "Assemble the Players",
		Completeness:      CompletenessFull,
		LibraryTopVisible: game.LibraryTopOwner,
		CastPermissions: []game.CastPermission{{
			Zone:             game.ZoneLibrary,
			TopOfLibraryOnly: true,
			CastOnly:         true,
			Filter:           game.PermissionFilter{CreatureOnly: true, PowerCapped: true, PowerAtMost: 2},
			PerType:          []string{game.PermissionTypeCreature},
			Label:            "Once each turn, cast a creature spell with power 2 or less from the top of your library (Assemble the Players)",
		}},
	})
}
