package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Experimental Frenzy — Enchantment {3}{R}:
//
//	"You may look at the top card of your library any time.
//	 You may play lands and cast spells from the top of your library.
//	 You can't play lands or cast spells from your hand.
//	 {3}{R}: Destroy this enchantment."
//
// ADR 0109 §4 (#1895). The first two sentences are Magus of the Future's
// and Bolas's Citadel's: the owner-visible top card
// (LibraryTopVisible) and a play-from-the-top permission (ADR 0066). The
// third is two statics read off the permanent's CONTROLLER, one on each
// gate: a CastRestriction and a LandPlayRestriction, both for the hand.
// CR 101.2 makes them beat every other permission to play from the hand,
// and neither reaches the command zone, the top of the library or a
// graveyard, so a commander can still be cast.
//
// The last line is Aether Storm's "Destroy this enchantment" body,
// activated by its controller only.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:          "3715e8b8-31df-499b-8a91-1bcd1199d6eb",
		Name:              "Experimental Frenzy",
		Completeness:      CompletenessFull,
		LibraryTopVisible: game.LibraryTopOwner,
		CastPermissions: []game.CastPermission{
			PlayFromTopOfYourLibrary(game.PermissionFilter{},
				"Play a land or cast a spell from the top of your library (Experimental Frenzy)"),
		},
		CastRestrictions: []game.CastRestriction{{
			Label: "You can't cast spells from your hand.",
			Forbids: func(q game.CastQuery) bool {
				return q.Source.Controller == q.Controller && q.FromZone == game.ZoneHand
			},
		}},
		LandPlayRestrictions: []game.LandPlayRestriction{{
			Label: "You can't play lands from your hand.",
			Forbids: func(q game.LandPlayQuery) bool {
				return q.Source.Controller == q.Player && q.FromZone == game.ZoneHand
			},
		}},
		Activated: []ActivatedAbility{{
			Label:  "{3}{R}: Destroy this enchantment.",
			Cost:   game.AbilityCost{Mana: "{3}{R}"},
			Effect: destroyThisPermanent(false),
		}},
	})
}
