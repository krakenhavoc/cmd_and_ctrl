package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Haakon, Stromgald Scourge — Legendary Creature — Zombie Knight {1}{B}{B},
// 3/3 (Magic Commander Collection, #5908):
//
//	"You may cast this card from your graveyard, but not from anywhere else.
//	 As long as Haakon is on the battlefield, you may cast Knight spells from
//	 your graveyard.
//	 When Haakon dies, you lose 2 life."
//
// The first line is Spec.CastsOnlyFromGraveyard beside CastableZones: the
// graveyard is open to the card and every other zone is refused under any
// permission (the aftermath zone rule, CR 702.127a's "can't be cast from any
// zone other than a graveyard"). The second is a standing cast permission over
// Knight cards in your graveyard that lives only while Haakon is on the
// battlefield. The death trigger is Haakon's own.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:               "8f089280-e5da-4fda-8bff-dea2971d0292",
		Name:                   "Haakon, Stromgald Scourge",
		Completeness:           CompletenessFull,
		CastableZones:          []game.ZoneKind{game.ZoneGraveyard},
		CastsOnlyFromGraveyard: true,
		CastPermissions: []game.CastPermission{{
			Zone:   game.ZoneGraveyard,
			Filter: game.PermissionFilter{Subtype: "Knight"},
			Label:  "Cast Knight spells from your graveyard (Haakon, Stromgald Scourge)",
		}},
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Haakon, Stromgald Scourge — you lose 2 life", Do(GainLife{Amount: -2})),
		},
	})
}
