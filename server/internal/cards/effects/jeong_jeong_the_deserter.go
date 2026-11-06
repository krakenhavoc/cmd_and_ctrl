package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Jeong Jeong, the Deserter — Legendary Creature — Human Rebel Ally
// {2}{R}, 2/3:
//
//	"Firebending 1 (Whenever this creature attacks, add {R}. This mana
//	 lasts until end of combat.)
//	 Exhaust — {3}: Put a +1/+1 counter on Jeong Jeong. When you next
//	 cast a Lesson spell this turn, copy it and you may choose new
//	 targets for the copy. (Activate each exhaust ability only once.)"
//
// Firebending 1 is CR 702.189a; exhaust is CR 702.177a.
//
// The second sentence is a CR 603.7 delayed trigger created by the
// ability's resolution, Doublecast's (delayed_on_event.go), watching for
// a Lesson spell (avB12YouNextCastALessonCondition) instead of an
// instant or sorcery. It is created even when Jeong Jeong has left the
// battlefield in response; only the counter needs him (CR 400.7). The
// copy is made even if the Lesson is countered in response, from
// last-known information, and is not cast.
func init() {
	Register(Spec{
		OracleID:     "2158c73d-421b-4c94-af06-dd89cb8d3126",
		Name:         "Jeong Jeong, the Deserter",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Jeong Jeong, the Deserter — firebending 1", Do(firebending(1))),
		},
		Activated: []ActivatedAbility{{
			Label:   "Exhaust — {3}: Put a +1/+1 counter on Jeong Jeong. When you next cast a Lesson spell this turn, copy it and you may choose new targets for the copy.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Exhaust: true,
			Cost:    ManaCost("{3}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if sourceIsStillThisPermanent(g, item) {
					if err := plusOneCountersOnThis(1)(g, item); err != nil {
						return err
					}
				}
				d := WhenYouNextCast("Jeong Jeong, the Deserter — copy that Lesson spell", game.CastFilter{}, copyTheSpellBody)
				d.Condition = avB12YouNextCastALessonCondition
				return d.Apply(NewContext(g, item))
			},
		}},
	})
}
