package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// The Holodeck — Legendary Artifact {7} (Star Trek, #4973):
//
//	"Once during each of your turns, you may play a permanent card you own
//	 from outside the game or from your graveyard."
//
// Two STANDING permissions (ADR 0066), one per place the card can come
// from, derived off the battlefield on every query, that share ONE use:
//
//   - Both hold the per-type budget of the single pseudo-type
//     PermissionTypePermanent (Serra Paragon's shape), and a derived
//     budget is kept per (holder, granting OBJECT, type), not per zone
//     (TurnTally.PermissionTypes, ADR 0066's 2026-11-05 amendment). So
//     the first land played or permanent spell cast through either
//     permission spends the use of both: a card from the graveyard closes
//     the sideboard for the turn, and the reverse. A second Holodeck is a
//     second object and brings a use of its own; a Holodeck that left and
//     came back is a new object (CR 400.7).
//   - Filter.PermanentOnly: a land (played, CR 305.1) or a permanent card,
//     never an instant or a sorcery. "Play" covers both, so neither
//     permission is cast-only, and a land played this way uses one of the
//     turn's land plays (CR 305.2).
//   - Timing is TimingYourTurnOnly: closed on other players' turns, and
//     each card keeps its own timing, so a land still needs your main
//     phase and an empty stack.
//   - "You own": a card goes to its owner's graveyard (CR 404.1), and the
//     outside-the-game container is the holder's own sideboard (CR
//     400.11a), so both permissions reach only the holder's own cards.
//     Filter.OwnedByHolder says it outright for a graveyard that a
//     Shaman's Trance has opened to every seat.
//   - The card is played for its own cost; the permission names no price.
//
// A database pool (no sideboard registered) names the card at the moment
// it is played, through the same outside_play action Wish uses: the
// action is open while any live permission over the container has a use
// left.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     theHolodeckOracleID,
		Name:         "The Holodeck",
		Completeness: CompletenessFull,
		CastPermissions: []game.CastPermission{
			theHolodeckPermission(game.ZoneOutside, "from outside the game"),
			theHolodeckPermission(game.ZoneGraveyard, "from your graveyard"),
		},
	})
}

const theHolodeckOracleID = "ecf28b5d-a719-47d4-8ad9-d24b6297ebde"

// theHolodeckPermission is one of the two halves. Both name the same
// budget type and come from the same object, which is what makes them one
// use.
func theHolodeckPermission(zone game.ZoneKind, from string) game.CastPermission {
	return game.CastPermission{
		Zone:    zone,
		Filter:  game.PermissionFilter{PermanentOnly: true, OwnedByHolder: true},
		PerType: []string{game.PermissionTypePermanent},
		Timing:  game.TimingYourTurnOnly,
		Label:   "Once during each of your turns, play a permanent card you own " + from + " (The Holodeck)",
	}
}
