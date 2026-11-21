package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// City of Death — Enchantment — Saga {2}{G} (Doctor Who, #5948):
//
//	"(As this Saga enters and after your draw step, add a lore counter.
//	 Sacrifice after VI.)
//	 I — Create a Treasure token.
//	 II, III, IV, V, VI — Create a token that's a copy of target non-Saga token
//	 you control."
//
// Chapter I makes a Treasure. Each later chapter targets a non-Saga token
// its controller controls, judged on announcement and again as it resolves
// (CR 608.2b), and makes a token that is a copy of it (CR 707.2), from its
// copiable values (or its last-known ones if it has left, CR 608.2h). The
// copies enter untapped and may themselves be copied by the next chapter.
//
// No simplification.
func init() {
	const name = "City of Death"
	triggers := []game.TriggeredAbility{
		ChapterTrigger(1, gapB92ChapterLabel(name, 1, "create a Treasure token"),
			Do(CreateToken{Template: TreasureToken(), N: 1})),
	}
	for n := 2; n <= 6; n++ {
		triggers = append(triggers, ChapterTriggerTargeting(n,
			gapB92ChapterLabel(name, n, "create a token that's a copy of target non-Saga token you control"),
			TargetPermanent("target non-Saga token you control", gapB92TokenYouControlNotASaga), gapB92CopyTheChosenToken))
	}
	Register(Spec{
		OracleID:     "313c6b29-c488-4b4b-af16-f82717f59a5d",
		Name:         name,
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1},
		Triggered:    triggers,
	})
}
