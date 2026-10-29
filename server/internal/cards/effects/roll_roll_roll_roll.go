package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Roll-Roll-Roll-Roll — Enchantment — Saga {2}{U} (The Hobbit):
//
//	"(As this Saga enters and after your draw step, add a lore counter.
//	 Sacrifice after IV.)
//	 I, II, III, IV — Exile up to one target creature or land you control.
//	 If you do, return it to the battlefield under its owner's control at
//	 the beginning of the next end step."
//
// Four chapters with one body: Oath of Teferi's delayed blink
// (exileTargetsThenScheduleReturn), over an "up to one" creature or land you
// control. With no target chosen nothing is exiled and nothing is scheduled.
// The card comes back as a new object under its owner's control at the next
// end step (CR 603.7), and a token exiled this way is gone for good
// (CR 111.8).
//
// No simplification.
func init() {
	const label = "Roll-Roll-Roll-Roll — exile up to one target creature or land you control; return it at the beginning of the next end step"
	chapters := make([]game.TriggeredAbility, 0, 4)
	for n := 1; n <= 4; n++ {
		chapters = append(chapters, ChapterTriggerTargeting(n, label,
			TargetPermanent("up to one target creature or land you control", Or(Creature(), Land()), YouControl()).WithCount(0, 1),
			func(g *game.Game, item *game.StackItem) error {
				return exileTargetsThenScheduleReturn(NewContext(g, item), "Roll-Roll-Roll-Roll — return the exiled permanent")
			}))
	}
	Register(Spec{
		OracleID:     "aa913409-7a7a-4e74-82ff-7ff8195585ad",
		Name:         "Roll-Roll-Roll-Roll",
		Completeness: CompletenessFull,
		Triggered:    chapters,
	})
}
