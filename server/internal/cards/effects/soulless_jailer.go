package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Soulless Jailer — {2} Artifact Creature — Phyrexian Golem 0/4:
//
//	"Permanent cards in graveyards can't enter the battlefield.
//	 Players can't cast noncreature spells from graveyards or exile."
//
// Sentence one is an entry restriction (#1985, entry_restriction.go):
// any permanent card, a land included, leaving a graveyard for the
// battlefield stays there (CR 101.2) — reanimation, a land played from a
// graveyard (the land-play gate refuses it before it is made), a
// creature returned by its own ability. A card in a library or in exile
// is not named, and a permanent cast from a graveyard is on the stack
// when it enters, so it resolves normally. Sentence two is the cast
// gate: a noncreature spell cast out of a graveyard or exile (flashback,
// an impulse draw, foretell) is refused, and a creature spell is not.
func init() {
	Register(Spec{
		OracleID:     "7866ab01-22f3-4d87-bdb0-c0ae0e91ffc9",
		Name:         "Soulless Jailer",
		Completeness: CompletenessFull,
		CastRestrictions: []game.CastRestriction{{
			Label: "Players can't cast noncreature spells from graveyards or exile.",
			Forbids: func(q game.CastQuery) bool {
				return (q.FromZone == game.ZoneGraveyard || q.FromZone == game.ZoneExile) && !q.Card.IsCreature()
			},
		}},
		EntryRestrictions: []game.EntryRestriction{
			CardsInZonesCantEnter("Permanent cards in graveyards can't enter the battlefield.",
				game.Card.IsPermanent, game.ZoneGraveyard),
		},
	})
}
