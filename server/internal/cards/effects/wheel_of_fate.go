package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wheel of Fate — Sorcery with no mana cost (Time Spiral, #5596):
//
//	"Suspend 4—{1}{R} (Rather than cast this card from your hand, pay
//	 {1}{R} and exile it with four time counters on it. At the
//	 beginning of your upkeep, remove a time counter. When the last is
//	 removed, you may cast it without paying its mana cost.)
//	 Each player discards their hand, then draws seven cards."
//
// Wheel of Fortune on a suspend timer. It has no mana cost, so it can
// never be cast from hand (CR 118.6); suspend (CR 702.62a) is the whole of
// its playability, and the last time counter's free cast resolves it.
// Every player discards their whole hand before anyone draws, the body
// shared with Wheel of Fortune.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "15e4938d-1cd3-4670-9de5-2dc10abe4c19",
		Name:         "Wheel of Fate",
		Completeness: CompletenessFull,
		SpecialActions: []game.SpecialAction{
			Suspend(4, "{1}{R}"),
		},
		OnResolve: wheelOfFortuneResolve,
	})
}
