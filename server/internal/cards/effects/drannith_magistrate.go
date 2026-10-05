package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Drannith Magistrate — Creature — Human Wizard {1}{W}, 1/3:
//
//	"Your opponents can't cast spells from anywhere other than their
//	 hands."
//
// A cast restriction keyed on the zone the spell is announced from,
// binding every player except the Magistrate's controller. The command
// zone is "anywhere other than their hands", so an opponent's commander
// cannot be cast while the Magistrate is out (the point of the card in
// Commander). Hand is the only exempt zone.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "aadd10d0-6dd0-4bdc-8d93-ff08e29a5863",
		Name:         "Drannith Magistrate",
		Completeness: CompletenessFull,
		CastRestrictions: []game.CastRestriction{{
			Label: "Your opponents can't cast spells from anywhere other than their hands.",
			Forbids: func(q game.CastQuery) bool {
				return q.Source.Controller != q.Controller && q.FromZone != game.ZoneHand
			},
		}},
	})
}
