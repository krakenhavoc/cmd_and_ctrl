package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Eye of Duskmantle — Creature — Eye {5}{B}{B} 3/8 (Murders at Karlov
// Manor Commander, #5429):
//
//	"Flying, lifelink
//	 You may play lands and cast spells from among cards in your graveyard
//	 you've surveilled this turn. If you cast a spell this way, you pay
//	 life equal to its mana value rather than paying its mana cost."
//
// A STANDING graveyard permission (ADR 0066) derived off the battlefield,
// Valgavoth, Terror Eater's shape with a different set of cards: the
// filter is PermissionFilter.SurveilledThisTurn, which reads the mark a
// surveil's graveyard leg leaves on each card it puts into the graveyard
// (Card.Surveilled, game/surveilled_cards.go). A card milled, discarded or
// already there does not qualify; a card kept on top was never in the
// graveyard; one that left the graveyard and came back is a new object
// (CR 400.7). The mark is written whether or not the Eye is out, so an Eye
// that arrives after the surveil still offers that turn's cards, and it
// lapses when the turn ends.
//
// The price is Bolas's Citadel's (LifeEqualToManaValue): an alternative
// cost, so the spell cannot be cast through the Eye for its mana cost, an
// {X} in it is 0, and a player without the life is not offered it
// (CR 119.4). A land is played, with the turn's land play and no life. The
// permission changes where the card may be played from, not when: a
// sorcery still needs its controller's main phase.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "04e7f739-fe50-440d-aa3d-e547295044a8",
		Name:            "Eye of Duskmantle",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "lifelink"},
		CastPermissions: []game.CastPermission{{
			Zone:                 game.ZoneGraveyard,
			Filter:               game.PermissionFilter{SurveilledThisTurn: true},
			AltCostKey:           eyeOfDuskmantleAltCost,
			LifeEqualToManaValue: true,
			Label:                "Pay life equal to its mana value (Eye of Duskmantle)",
		}},
	})
}

// eyeOfDuskmantleAltCost is the alternative-cost key a cast through the Eye
// names.
const eyeOfDuskmantleAltCost = "eye_of_duskmantle"
