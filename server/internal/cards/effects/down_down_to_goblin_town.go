package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Down, Down to Goblin-town — Enchantment — Saga {2}{B}:
//
//	"I — Target opponent reveals their hand. You choose a nonland card
//	     from it. That player discards that card.
//	 II — Amass Goblins 1.
//	 III, IV — Target opponent loses 1 life and you gain 1 life."
//
// Four chapter abilities (CR 714.2b, 714.2c: "III, IV" is two
// abilities with the same text), so the Saga is sacrificed after IV.
// Chapter I is Thoughtseize's pick (ADR 0116), chosen by the Saga's
// controller (CR 113.8). Chapter II is the shared Amass (CR 701.47a):
// a 0/0 black Goblin Army token if you control no Army, then a +1/+1
// counter, and the chosen Army becomes a Goblin. Chapters III and IV
// are the shared one-point drain.
//
// No simplification.
func init() {
	const name = "Down, Down to Goblin-town"
	drain := func(n int) game.TriggeredAbility {
		return ChapterTriggerTargeting(n, SagaChapterLabel(name, n, "target opponent loses 1 life and you gain 1 life"),
			TargetPlayer("target opponent", Opponent()),
			drainTargetOpponentOne)
	}
	Register(Spec{
		OracleID:     "e01860cd-0aa1-435a-9ef9-ee412bc458cd",
		Name:         name,
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			ChapterTriggerTargeting(1, SagaChapterLabel(name, 1, "target opponent reveals their hand"),
				TargetPlayer("target opponent", Opponent()),
				TargetRevealsYouChooseDiscardAbility(Nonland(), "nonland card")),
			ChapterTrigger(2, SagaChapterLabel(name, 2, "amass Goblins 1"),
				Do(Amass{Subtype: "Goblin", N: 1})),
			drain(3),
			drain(4),
		},
	})
}
