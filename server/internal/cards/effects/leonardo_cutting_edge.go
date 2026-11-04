package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Leonardo, Cutting Edge — Legendary Creature — Mutant Ninja Turtle
// {1}{W}, 1/1 (TMT, #4457):
//
//	"Sneak {W} (… He enters tapped and attacking.)
//	 Lifelink
//	 Whenever you gain life, put a +1/+1 counter on Leonardo."
//
// Sneak is Sneak(cost) (CR 702.190); the trigger is Ajani's Pridemate's,
// once per life-gain event, his own lifelink included.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "933fe6e5-8d00-4411-aa54-7780382c1ea6",
		Name:             "Leonardo, Cutting Edge",
		Completeness:     CompletenessFull,
		PrintedKeywords:  []string{"lifelink"},
		AlternativeCosts: []game.AlternativeCost{Sneak("{W}")},
		Triggered: []game.TriggeredAbility{
			WheneverYouGainLife("Leonardo, Cutting Edge — put a +1/+1 counter on him", Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
