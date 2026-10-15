package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kunoros, Hound of Athreos — {1}{W}{B} Legendary Creature — Dog 3/3:
//
//	"Vigilance, menace, lifelink
//	 Creature cards in graveyards can't enter the battlefield.
//	 Players can't cast spells from graveyards."
//
// The three keywords are printed and enforced. Sentence two is an entry
// restriction (#1985, entry_restriction.go): a creature card leaving a
// graveyard for the battlefield stays there (CR 101.2), whoever is
// returning it — Kunoros's own controller included — and a creature land
// (Dryad Arbor) played from a graveyard is refused before it is made.
// Sentence three is the cast gate, about the graveyard alone.
func init() {
	Register(Spec{
		OracleID:        "d297ce39-da98-4351-be9d-24f4d09bee7f",
		Name:            "Kunoros, Hound of Athreos",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance", "menace", "lifelink"},
		CastRestrictions: []game.CastRestriction{
			PlayersCantCastFrom("Players can't cast spells from graveyards.", game.ZoneGraveyard),
		},
		EntryRestrictions: []game.EntryRestriction{
			CardsInZonesCantEnter("Creature cards in graveyards can't enter the battlefield.",
				game.Card.IsCreature, game.ZoneGraveyard),
		},
	})
}
