package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Guru Pathik — Legendary Creature — Human Monk Ally {2}{G/U}{G/U}, 2/4
// (Avatar: The Last Airbender, #4459):
//
//	"When Guru Pathik enters, look at the top five cards of your library.
//	 You may reveal a Lesson, Saga, or Shrine card from among them and put
//	 it into your hand. Put the rest on the bottom of your library in a
//	 random order.
//	 Whenever you cast a Lesson, Saga, or Shrine spell, put a +1/+1 counter
//	 on another target creature you control."
//
// The enters trigger is Horn of the Mark's sentence
// (LookAtTopThenMayTakeToHand). The cast trigger reads the spell's
// subtypes as it is cast (CR 601.2i) and targets a creature you control
// other than Guru Pathik, chosen as it goes on the stack; with no such
// creature it is removed (CR 603.3d).
//
// No simplification.
func init() {
	lessonSagaOrShrine := AnySubtype("Lesson", "Saga", "Shrine")
	Register(Spec{
		OracleID:     "4dcd7b78-ce5a-43eb-a16d-3b9ea411795d",
		Name:         "Guru Pathik",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Guru Pathik — look at the top five cards of your library",
				LookAtTopThenMayTakeToHand(5, lessonSagaOrShrine, 1,
					"Guru Pathik — reveal a Lesson, Saga, or Shrine card and put it into your hand")),
			Targeting(WheneverYouCast(lessonSagaOrShrine,
				"Guru Pathik — put a +1/+1 counter on another target creature you control", PlusOneCounterOnTheTarget),
				Another(TargetCreature("another target creature you control", YouControl()))),
		},
	})
}
