package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Riftstone Portal — Land (Judgment):
//
//	"{T}: Add {C}.
//	 As long as this card is in your graveyard, lands you control have "{T}:
//	 Add {G} or {W}.""
//
// Brawn's graveyard static (CR 113.6c, #1221) giving the granted-mana bundle
// of Joiner Adept to every land its owner controls. "You" is the card's owner
// (CR 108.4), and the grant does not apply while the Portal is on the
// battlefield: the declared zone is the graveyard alone, and the Portal's own
// mana ability is the printed {C}.
//
// No simplification.
func init() {
	const grant = "riftstone-portal/green-or-white"
	anthem := effects.GrantAbilities(effects.LandsYouControl, grant)
	anthem.Label = "As long as this card is in your graveyard, lands you control have \"{T}: Add {G} or {W}.\""
	anthem.Zones = []game.ZoneKind{game.ZoneGraveyard}
	effects.Register(effects.Spec{
		OracleID:     "8d7e05ba-5406-4d5e-bb8f-a4a6f3b0eaa7",
		Name:         "Riftstone Portal",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Grants: []effects.AbilityGrant{effects.TapForManaGrant(grant, "{G|W}", "Add {G} or {W}", "{T}: Add {G} or {W}.")},
		Static: []game.StaticAbility{anthem},
	})
}
