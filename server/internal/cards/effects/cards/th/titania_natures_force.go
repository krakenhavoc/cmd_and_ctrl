package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Titania, Nature's Force:
//
//	"You may play Forests from your graveyard.
//	 Whenever a Forest you control enters, create a 5/3 green Elemental
//	 creature token.
//	 Whenever an Elemental you control dies, you may mill three cards."
//
// The permission is Crucible of Worlds' play permission narrowed to the
// Forest land type; a land played this way uses the land drop (CR 305.2).
// The dies trigger leaves the battlefield with last-known information, so
// Titania herself (an Elemental) triggers it (CR 603.10a, 608.2h).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "46c3a69a-4984-40b0-905f-4b0762c7ad72",
		Name:         "Titania, Nature's Force",
		Completeness: effects.CompletenessFull,
		CastPermissions: []game.CastPermission{{
			Zone:   game.ZoneGraveyard,
			Filter: game.PermissionFilter{LandsOnly: true, Subtype: "Forest"},
			Label:  "Play a Forest from your graveyard (Titania, Nature's Force)",
		}},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, effects.Cmdr2S14SubtypeLandEntersUnderYourControl("Forest"),
				"Titania, Nature's Force — create a 5/3 green Elemental creature token",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("5/3 green Elemental"), N: 1})),
			effects.Optional(effects.On(game.EventLTB, effects.Cmdr2S14ElementalYouControlDied,
				"Titania, Nature's Force — you may mill three cards",
				effects.Do(effects.MillCards{N: 3})),
				"Titania, Nature's Force — mill three cards?"),
		},
	})
}
