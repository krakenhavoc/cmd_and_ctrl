package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Balance of Terror — Enchantment — Saga {2}{W} (Star Trek Commander,
// #4974):
//
//	"(As this Saga enters and after your draw step, add a lore counter.
//	 Sacrifice after III.)
//	 I — Destroy up to one other target artifact or enchantment.
//	 II, III — Exile up to one other target permanent you control.
//	 Return that card to the battlefield under its owner's control at
//	 the beginning of the next end step."
//
// "Other" is this Saga as an object (Another, CR 109.1). Chapter I may
// target nothing. "II, III" is two chapter abilities with one text
// (CR 714.2c), so it is two entries. Each is Test of Priorities' delayed
// blink: the card returns under its OWNER's control at the next end step
// (a CR 603.7 delayed trigger), as a new object (CR 400.7). A target
// that left in response is not exiled and nothing returns.
func init() {
	const name = "Balance of Terror"
	blink := func(n int) game.TriggeredAbility {
		label := SagaChapterLabel(name, n, "exile up to one other target permanent you control; it returns at the next end step")
		return ChapterTriggerTargeting(n, label,
			Another(TargetPermanent("up to one other target permanent you control", YouControl()).WithCount(0, 1)),
			func(g *game.Game, item *game.StackItem) error {
				return exileTargetsThenScheduleReturn(NewContext(g, item), "Balance of Terror — return the exiled card")
			})
	}
	Register(Spec{
		OracleID:     "0b035f3c-ba21-4a73-b4df-11a7e8c19056",
		Name:         name,
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			ChapterTriggerTargeting(1, SagaChapterLabel(name, 1, "destroy up to one other target artifact or enchantment"),
				Another(TargetPermanent("up to one other target artifact or enchantment", Or(Artifact(), Enchantment())).WithCount(0, 1)),
				destroyFirstLegalTarget),
			blink(2),
			blink(3),
		},
	})
}
