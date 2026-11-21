package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Barbara Wright — Legendary Creature — Human Advisor {1}{W}, 1/3 (Doctor
// Who, #5948):
//
//	"History Teacher — Sagas you control have read ahead. (As a Saga enters,
//	 choose a chapter and start with that many lore counters. Skipped chapters
//	 don't trigger.)
//	 Doctor's companion"
//
// A layer-6 keyword grant over the Sagas the controller controls (KeywordGrant).
// Read ahead is an entry choice (CR 702.155a), and the engine's entry
// look-ahead reads the entering Saga's abilities through the layers with this
// grant applied (CR 614.12), so a Saga that enters while Barbara is on the
// battlefield asks the chapter. Doctor's companion is a deck-construction
// rule; "History Teacher" is an ability word (CR 207.2c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0e61a062-f13a-4958-91de-909650c662a8",
		Name:         "Barbara Wright",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			Labelled(KeywordGrant(func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.Controller == source.Controller && game.IsSaga(*target)
			}, game.KeywordReadAhead), "Sagas you control have read ahead."),
		},
	})
}
