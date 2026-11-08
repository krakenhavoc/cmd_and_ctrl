package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Summon: Fat Chocobo — Enchantment Creature — Saga Bird {4}{G}, 4/4
// (Final Fantasy, #4462):
//
//	"(As this Saga enters and after your draw step, add a lore counter.
//	 Sacrifice after IV.)
//	 I — Wark — Create a 2/2 green Bird creature token with 'Whenever a
//	 land you control enters, this token gets +1/+0 until end of turn.'
//	 II, III, IV — Kerplunk — Creatures you control gain trample until
//	 end of turn."
//
// A Saga creature. Chapter I makes the set's landfall Bird (Gysahl
// Greens' token). Chapters II–IV are Crash Through's grant: the
// creatures you control as each resolves gain trample (CR 611.2c).
//
// No simplification.
func init() {
	const name = "Summon: Fat Chocobo"
	kerplunk := Do(GrantKeywordUntilEOT{
		Match:    And(Creature(), YouControl()),
		Keywords: []string{"trample"},
		Label:    name + " — creatures you control gain trample",
	})
	Register(Spec{
		OracleID:     "b5d2f8e7-89d3-4015-ab99-ba4d66918e37",
		Name:         name,
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			ChapterTrigger(1, SagaChapterLabel(name, 1, "Wark — create a 2/2 green Bird token with landfall +1/+0"),
				Do(CreateToken{Template: chocoboBirdToken(), N: 1})),
			ChapterTrigger(2, SagaChapterLabel(name, 2, "Kerplunk — creatures you control gain trample until end of turn"), kerplunk),
			ChapterTrigger(3, SagaChapterLabel(name, 3, "Kerplunk — creatures you control gain trample until end of turn"), kerplunk),
			ChapterTrigger(4, SagaChapterLabel(name, 4, "Kerplunk — creatures you control gain trample until end of turn"), kerplunk),
		},
	})
}
