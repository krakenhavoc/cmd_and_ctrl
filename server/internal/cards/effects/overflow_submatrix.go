package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Overflow Submatrix — Artifact Creature — Borg {2}{B}, 3/3 (Star Trek,
// #4973):
//
//	"Once during each of your turns, you may cast an artifact creature
//	 spell from your graveyard. If you cast a spell this way, that
//	 artifact creature enters with a finality counter on it."
//
// A STANDING graveyard permission (ADR 0066) derived off the battlefield,
// Gisa and Geralf's shape: a budget of one creature cast each of your
// turns (PerType, spent per source and reset each turn), closed off your
// turn by TimingYourTurnOnly, which leaves each spell's own timing in
// force. "Artifact creature spell" is judged on the face being cast
// (CR 601.3e). "That artifact creature enters with a finality counter" is
// the permission's rider (CastPermission.EntersWithCounterKind, #4679),
// put on through the CR 614 pipeline; the finality counter then exiles it
// instead of letting it die (CR 122.1h). A card cast from your graveyard
// by its own text is not cast "this way" and spends nothing.
//
// No simplification.
func init() {
	const name = "Overflow Submatrix"
	Register(Spec{
		OracleID:     "523c244b-7292-40c7-8b12-27106fffe63f",
		Name:         name,
		Completeness: CompletenessFull,
		CastPermissions: []game.CastPermission{{
			Zone:                  game.ZoneGraveyard,
			CastOnly:              true,
			Filter:                game.PermissionFilter{CreatureOnly: true, CardType: "artifact"},
			PerType:               []string{game.PermissionTypeCreature},
			Timing:                game.TimingYourTurnOnly,
			EntersWithCounterKind: game.CounterFinality,
			EntersWithCounters:    1,
			Label:                 "Once during each of your turns, cast an artifact creature spell from your graveyard; it enters with a finality counter (" + name + ")",
		}},
	})
}
