package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// The City on the Edge of Forever — Enchantment — Saga {3}{W}{B} (Star
// Trek, #4973):
//
//	"(As this Saga enters and after your draw step, add a lore counter.
//	 Sacrifice after III.)
//	 I — All artifacts and creatures phase out. Create a 0/1 white Human
//	 creature token with "Permanents can't phase in." (Treat phased-out
//	 permanents and anything attached to them as though they don't
//	 exist.)
//	 II — You draw two cards and each player loses 2 life.
//	 III — Sacrifice a creature token. If you do, you gain 5 life."
//
// Chapter I phases every artifact and creature out at once (CR 702.26a,
// attachments with their hosts, CR 702.26g), then creates the Human. Its
// line is the PermanentsCantPhaseIn world rule: while it is on the
// battlefield nothing phases in, at an untap step or anywhere else; once
// it is gone the permanents phase in at the next untap step of the
// player they phased out under. Chapter II draws, then each player, you
// included, loses 2 life. Chapter III asks for a creature token you
// control as it resolves (the Human is one) and gains 5 life only when
// one was sacrificed.
//
// No simplification.
func init() {
	const name = "The City on the Edge of Forever"
	Register(Spec{
		OracleID:     "47c9ef8c-7ba7-4702-bd8c-5fedc59c8166",
		Name:         name,
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Draws: 2, CreatureTokens: 1, Tokens: 1},
		Triggered: []game.TriggeredAbility{
			ChapterTrigger(1, SagaChapterLabel(name, 1, "all artifacts and creatures phase out; create a 0/1 Human that stops phasing in"), trekB23CityChapterOne),
			ChapterTrigger(2, SagaChapterLabel(name, 2, "you draw two cards and each player loses 2 life"),
				Do(DrawCards{N: 2}, Instruction(eachPlayerLosesTwoLife))),
			ChapterTrigger(3, SagaChapterLabel(name, 3, "sacrifice a creature token; if you do, you gain 5 life"), trekB23CityChapterThree),
		},
	})
}
