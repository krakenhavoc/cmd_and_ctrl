package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Thunder of Unity — Enchantment — Saga {R}{W}{B} (Tarkir: Dragonstorm,
// #4463):
//
//	"(As this Saga enters and after your draw step, add a lore counter.
//	 Sacrifice after III.)
//	 I — You draw two cards and you lose 2 life.
//	 II, III — Whenever a creature you control enters this turn, each
//	 opponent loses 1 life and you gain 1 life."
//
// Chapter I is the draw and the life loss. Chapters II and III are The
// Naked Time's repeating delayed trigger (CR 603.7b) over "each creature
// you control enters": it triggers for every creature that enters under
// your control for the rest of the turn, the Saga leaving or not, and each
// instance drains each opponent for 1 (every opponent still in the game)
// and gains you 1.
//
// No simplification.
func init() {
	const name = "Thunder of Unity"
	drain := func(g *game.Game, item *game.StackItem) error {
		return DelayedOnEvent{
			Label:     "Thunder of Unity — each opponent loses 1 life and you gain 1 life",
			On:        []game.EventKind{game.EventETB},
			Condition: eachCreatureYouControlEntersCondition,
			Body:      thunderOfUnityDrainBody,
			Repeats:   true,
		}.Apply(NewContext(g, item))
	}
	Register(Spec{
		OracleID:     "8419d1d5-bb0e-4a2d-bb6f-67957d035dde",
		Name:         name,
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			ChapterTrigger(1, SagaChapterLabel(name, 1, "you draw two cards and you lose 2 life"),
				Do(DrawCards{N: 2}, GainLife{Amount: -2})),
			ChapterTrigger(2, SagaChapterLabel(name, 2, "whenever a creature you control enters this turn, each opponent loses 1 life and you gain 1 life"), drain),
			ChapterTrigger(3, SagaChapterLabel(name, 3, "whenever a creature you control enters this turn, each opponent loses 1 life and you gain 1 life"), drain),
		},
	})
}

// thunderOfUnityDrainBody is the delayed trigger's body. An on-disk key:
// never renamed, never reused.
var thunderOfUnityDrainBody = game.SimpleDelayedBody("thunder-of-unity/drain",
	Do(EachOpponentLosesLife{N: 1}, GainLife{Amount: 1}))
