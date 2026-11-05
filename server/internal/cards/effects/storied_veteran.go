package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Storied Veteran — Creature — Klingon Advisor {B}, 2/1 (Star Trek,
// #4973):
//
//	"This creature can't block.
//	 You may cast this card from your graveyard. If you do, it enters
//	 with a finality counter on it."
//
// The graveyard is a zone the card's own text opens (Spec.CastableZones),
// at its ordinary cost and timing. "If you do, it enters with a finality
// counter" is a CR 614.1c clause read off the resolving spell's cast
// record (CountersIfCastFrom): cast from your hand it enters with none,
// and put onto the battlefield without being cast it enters with none.
// The finality counter exiles it instead of letting it die (CR 122.1h).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "b1012ac3-77f8-4384-b947-84c094d6b59e",
		Name:          "Storied Veteran",
		Completeness:  CompletenessFull,
		Static:        []game.StaticAbility{RestrictSelf(game.CantBlock)},
		CastableZones: []game.ZoneKind{game.ZoneGraveyard},
		EntersWithCountersFromCast: []game.EntryCountersFromCast{
			CountersIfCastFrom(game.ZoneGraveyard, game.CounterFinality, 1),
		},
	})
}
