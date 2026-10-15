package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	attack := WheneverThisAttacks("Turtle Van — put a +1/+1 counter on target creature that crewed it, then double if it is a Mutant, Ninja or Turtle", turtleVanCounter)
	attack.Targets = CrewedThis(TargetCreature("target creature that crewed it this turn"))

	Register(Spec{
		OracleID:     "aeb9a2e6-ad11-4e2c-81d0-caf279d31ece",
		Name:         "Turtle Van",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "Crew 1",
			Cost:    CrewCost(1),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect:  CrewEffect("Turtle Van"),
		}},
		Triggered: []game.TriggeredAbility{attack},
	})
}

func turtleVanCounter(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	return g.AddCounterThenForEffect(id, game.CounterPlusOne, 1, func(g *game.Game, _ int) error {
		c, found := g.LookupCardForEffect(id)
		if !found || !onBattlefield(g, id) {
			return nil
		}
		if !c.HasSubtype("Mutant") && !c.HasSubtype("Ninja") && !c.HasSubtype("Turtle") {
			return nil
		}
		n := c.Counters[game.CounterPlusOne]
		if n <= 0 {
			return nil
		}
		return AddCounter{Target: id, Kind: game.CounterPlusOne, N: n}.Apply(NewContext(g, item))
	})
}
