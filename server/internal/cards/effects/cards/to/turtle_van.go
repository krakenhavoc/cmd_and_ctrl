package cardsto

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Turtle Van — {3} Artifact — Vehicle, 4/4:
//
//	"Whenever this Vehicle attacks, put a +1/+1 counter on target
//	 creature that crewed it this turn. Then if that creature is a
//	 Mutant, Ninja, or Turtle, double the number of +1/+1 counters on it.
//	 Crew 1"
//
// The target clause is CrewedThis (ADR 0186 section 3): the creatures
// tapped to pay THIS Van's crew cost this turn, read from the turn's crew
// records by the Van's object and still on the battlefield as the same
// objects. A creature that crewed another Vehicle is not offered, a crewer
// that left is not, and with none left the trigger is removed for want of
// a target (CR 603.3d).
//
// The doubling reads the count AFTER the first counter has landed (the
// Invigorating Surge order), as a second placement, so a counter doubler
// sees both. "Is a Mutant, Ninja, or Turtle" reads the creature's
// effective subtypes after the counter is placed, so a changeling counts.
//
// No simplification.
func init() {
	attack := effects.WheneverThisAttacks("Turtle Van — put a +1/+1 counter on target creature that crewed it, then double if it is a Mutant, Ninja or Turtle", turtleVanCounter)
	attack.Targets = effects.CrewedThis(effects.TargetCreature("target creature that crewed it this turn"))

	effects.Register(effects.Spec{
		OracleID:     "aeb9a2e6-ad11-4e2c-81d0-caf279d31ece",
		Name:         "Turtle Van",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "Crew 1",
			Cost:    effects.CrewCost(1),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect:  effects.CrewEffect("Turtle Van"),
		}},
		Triggered: []game.TriggeredAbility{attack},
	})
}

func turtleVanCounter(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	id, ok := effects.B16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	then := ctx.ItemTail(turtleVanDouble)
	then.Carry = []uuid.UUID{id}
	return g.AddCounterByKeyedForEffect(uuid.Nil, id, game.CounterPlusOne, 1, then)
}

// turtleVanDouble is "if it's a Mutant, Ninja or Turtle, double the number of
// +1/+1 counters on it" once the counter has landed; Carry[0] is the creature.
// "Mutant, Ninja or Turtle" are words of the source's text (CR 612.2, #3984),
// read when the tail runs, as the asking item's text. An on-disk key: never
// renamed, never reused.
var turtleVanDouble = effects.CardTail("turtle-van/double", func(ctx *effects.Context, o game.TailOutcome) error {
	id := effects.CarryFirst(o.Tail)
	mutant, ninja, turtle := ctx.TextSubtype("Mutant"), ctx.TextSubtype("Ninja"), ctx.TextSubtype("Turtle")
	c, found := ctx.Game.LookupCardForEffect(id)
	if !found || !effects.OnBattlefield(ctx.Game, id) {
		return nil
	}
	if !c.HasSubtype(mutant) && !c.HasSubtype(ninja) && !c.HasSubtype(turtle) {
		return nil
	}
	return effects.B23DoublePlusOneCountersOn(ctx, id)
})
