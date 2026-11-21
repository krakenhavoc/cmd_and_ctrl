package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Restricted Office // Lecture Hall — Enchantment — Room (CR 709.5):
//
//	Restricted Office {2}{W}{W}: "When you unlock this door, destroy all
//	creatures with power 3 or greater."
//	Lecture Hall {5}{U}{U}: "Other permanents you control have hexproof."
func init() {
	effects.Register(effects.Room(effects.RoomSpec{
		OracleID:     "b36286fa-4007-4276-af9c-d8f0da9e94d1",
		Name:         "Restricted Office // Lecture Hall",
		Completeness: effects.CompletenessFull,
		Left: effects.Door{Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.WhenYouUnlockThisDoor(game.DoorLeft, "Restricted Office — destroy all creatures with power 3 or greater",
				effects.Do(effects.DestroyAllMatching{Match: effects.And(effects.Creature(), effects.PowerGE(3))})), game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy, Partial: true}}),
		}},
		Right: effects.Door{Static: []game.StaticAbility{effects.B16GrantKeywords(
			func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.InstanceID != source.InstanceID && target.Controller == source.Controller
			}, "hexproof")}},
	}))
}
