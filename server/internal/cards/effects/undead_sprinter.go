package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Undead Sprinter — Creature — Zombie {B}{R}, 2/2 (Duskmourn: House of
// Horror, #4466):
//
//	"Trample, haste
//	 You may cast this card from your graveyard if a non-Zombie creature
//	 died this turn. If you do, this creature enters with a +1/+1
//	 counter on it."
//
// The graveyard is a cast zone the card opens for itself only while the
// condition holds (Spec.CastableZonesIf, Gravecrawler's shape), asked
// whenever the cast is offered, at the card's ordinary cost and timing. The
// counter is a CR 614.1c clause read off the resolving spell's cast record
// (CountersIfCastFrom, Storied Veteran's): cast from hand it enters with
// none, and a card put onto the battlefield without being cast enters with
// none. A countered or dead Sprinter returns to the graveyard and may be
// cast that way again.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "3416ac84-c5ef-44be-894f-ab3b89e592ce",
		Name:            "Undead Sprinter",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample", "haste"},
		CastableZonesIf: []game.ConditionalCastZone{{Zone: game.ZoneGraveyard, Cond: stdB26NonZombieCreatureDied}},
		EntersWithCountersFromCast: []game.EntryCountersFromCast{
			CountersIfCastFrom(game.ZoneGraveyard, game.CounterPlusOne, 1),
		},
	})
}
