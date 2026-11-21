package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Cathedral Acolyte — Creature — Human Cleric {1}{G}, 1/2 (Fallout,
// #5948):
//
//	"Each creature you control with a counter on it has ward {1}.
//	 (Whenever it becomes the target of a spell or ability an opponent
//	 controls, counter it unless that player pays {1}.)
//	 {T}: Put a +1/+1 counter on target creature that entered this turn."
//
// The ward is Bronze Guardian's grant (WardGranted) over the creatures its
// controller controls that carry a counter of any kind as the targeting
// happens (the Acolyte itself once it has one). The activated ability's
// target is a creature, any controller's, that entered the battlefield this
// turn (the turn's entry tally, CR 400.7: a creature that left and came
// back entered again); the counter goes on only if it still qualifies as the
// ability resolves (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "85fa4ed3-14e3-4eb0-b31b-62e2088321d0",
		Name:         "Cathedral Acolyte",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WardGranted(WardMana("{1}"), "Cathedral Acolyte — creatures you control with a counter on them have ward {1}",
				func(target *game.Card, g *game.Game, source *game.Card) bool {
					return target.IsCreature() && target.Controller == source.Controller && b64HasAnyCounter()(g, uuid.Nil, *target)
				}),
		},
		Activated: []ActivatedAbility{{
			Label:   "{T}: Put a +1/+1 counter on target creature that entered this turn.",
			Cost:    TapCost(),
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Targets: TargetCreature("target creature that entered this turn",
				func(g *game.Game, _ uuid.UUID, c game.Card) bool { return g.EnteredThisTurn(c.InstanceID) }),
			Effect: putACounterOnTheTarget(game.CounterPlusOne),
		}},
	})
}
