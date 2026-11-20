package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cutthroat il-Dal — Creature — Human Rogue {3}{B}, 4/1 (Future Sight, #5594):
//
//	"Hellbent — This creature has shadow as long as you have no cards
//	 in hand. (It can block or be blocked by only creatures with
//	 shadow.)"
//
// A layer 6 shadow grant (CR 613.1f) that is on while its controller's
// hand is empty, read through the hand size the layers watch
// (DependsOnHandSize), so drawing a card takes shadow away at once.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c70526a6-9071-4f34-b5fb-9d075021f667",
		Name:         "Cutthroat il-Dal",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{futB1HellbentKeyword("shadow")},
	})
}
