package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// MJ, Rising Star — Legendary Creature — Human Performer {2}{W}, 2/3
// (Marvel Super Heroes Commander, #4455):
//
//	"Vigilance
//	 Whenever you gain life, put a +1/+1 counter on MJ."
//
// Vigilance rides PrintedKeywords. The trigger is Archangel of Thune's
// lifegain condition (every lifegain path, lifelink included, raises one
// positive life-change event on the controller; one event is one trigger),
// and the counter goes on MJ herself, if she is still the same object.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "06f647a7-ee53-42bd-8c73-53a0c6ede34a",
		Name:            "MJ, Rising Star",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			WheneverYouGainLife("MJ, Rising Star — a +1/+1 counter on MJ", Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
