package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Restricted Office // Lecture Hall — Enchantment — Room (CR 709.5):
//
//	Restricted Office {2}{W}{W}: "When you unlock this door, destroy all
//	creatures with power 3 or greater."
//	Lecture Hall {5}{U}{U}: "Other permanents you control have hexproof."
func init() {
	Register(Room(RoomSpec{
		OracleID:     "b36286fa-4007-4276-af9c-d8f0da9e94d1",
		Name:         "Restricted Office // Lecture Hall",
		Completeness: CompletenessFull,
		Left: Door{Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(WhenYouUnlockThisDoor(game.DoorLeft, "Restricted Office — destroy all creatures with power 3 or greater",
				Do(DestroyAllMatching{Match: And(Creature(), PowerGE(3))})), game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy, Partial: true}}),
		}},
		Right: Door{Static: []game.StaticAbility{b16GrantKeywords(
			func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.InstanceID != source.InstanceID && target.Controller == source.Controller
			}, "hexproof")}},
	}))
}
