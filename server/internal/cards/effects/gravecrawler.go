package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Gravecrawler — Creature — Zombie {B}, 2/1:
//
//	"This creature can't block.
//	 You may cast this card from your graveyard as long as you control
//	 a Zombie."
//
// The first card on Spec.CastableZonesIf that is not a speed card
// (2026-10-13 amendment to ADR 0066): the graveyard is a cast zone for
// it only while its owner controls a Zombie. "A Zombie" reads effective
// subtypes, so a changeling or a creature turned into a Zombie counts,
// and Gravecrawler itself does not (it is in the graveyard, not on the
// battlefield).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "09ff28b1-b6c9-48e6-b12e-2f0e644f709f",
		Name:            "Gravecrawler",
		Completeness:    CompletenessFull,
		Static:          []game.StaticAbility{RestrictSelf(game.CantBlock)},
		CastableZonesIf: []game.ConditionalCastZone{{Zone: game.ZoneGraveyard, Cond: ownerControlsAZombie}},
	})
}

// ownerControlsAZombie is "as long as you control a Zombie", asked of
// the card's owner: a pure read of the battlefield.
func ownerControlsAZombie(g *game.Game, owner uuid.UUID, _ game.Card) bool {
	return ControlsA("Zombie")(g, owner)
}
