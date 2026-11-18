package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Squee, the Immortal — {1}{R}{R} Legendary Creature — Goblin, 2/1:
//
//	"You may cast this card from your graveyard or from exile."
//
// The card's own text opens both zones for its owner (CR 601.2a),
// unconditionally: the graveyard through CastableZones, exile through the
// CastsItselfFromExile declaration Khan, Vengeful Heart introduced (ADR
// 0066 amendment 2026-11-05). A Squee exiled face down has no text and is
// not opened. The cast is an ordinary one at the card's own cost.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:             "481fb71a-11b9-462e-a4da-ca35ffd1bd22",
		Name:                 "Squee, the Immortal",
		Completeness:         CompletenessFull,
		CastableZones:        []game.ZoneKind{game.ZoneGraveyard, game.ZoneExile},
		CastsItselfFromExile: true,
	})
}
