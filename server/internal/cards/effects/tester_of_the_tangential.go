package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tester of the Tangential — Creature — Djinn Wizard {1}{U}, 1/1 (Secrets of
// Strixhaven, #4456):
//
//	"Increment (Whenever you cast a spell, if the amount of mana you spent
//	 is greater than this creature's power or toughness, put a +1/+1
//	 counter on this creature.)
//	 At the beginning of combat on your turn, you may pay {X}. When you do,
//	 move X +1/+1 counters from this creature onto another target
//	 creature."
//
// Increment is the keyword helper (CR 702.191). The combat trigger is
// Electro's "you may pay {X}. When you do" (MayPayX): X is chosen as the
// trigger resolves (CR 608.2d), paid through the ordinary "you may pay"
// prompt, and paying nothing is declining. Paying makes a CR 603.12
// reflexive trigger that picks "another target creature" (any player's) as
// it goes on the stack and carries X.
//
// Moving counters (CR 122.5) takes them off this creature and puts them on
// the target: no more than the creature has, and nothing if it has left the
// battlefield or come back as a new object, or if the target is gone. A
// counter that can't be removed isn't moved.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "a30a2b96-4a8c-4d18-b276-5005b8aff7ed",
		Name:            "Tester of the Tangential",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{KeywordIncrement},
		ReadsManaSpent:  game.ReadsYourCasts,
		Triggered: []game.TriggeredAbility{
			Increment(),
			AtBeginningOfYourCombat("Tester of the Tangential — you may pay {X}; move X +1/+1 counters onto another target creature",
				testerOfTheTangentialPay),
		},
	})
}

// testerOfTheTangentialPaid is the paid half, keyed: X rides the prompt.
var testerOfTheTangentialPaid = PayUnlessBranches("pay-unless/tester-of-the-tangential-move",
	func(ctx *Context, r game.PayUnlessAnswer) error {
		if r.Number <= 0 {
			return nil
		}
		return ReflexiveTrigger{
			Label:  "Tester of the Tangential — move X +1/+1 counters onto another target creature",
			Body:   testerOfTheTangentialMoveBody,
			Params: game.EffectParams{Amount: r.Number},
		}.Apply(ctx)
	}, nil)

// testerOfTheTangentialMoveBody is the reflexive trigger: "move X +1/+1
// counters from this creature onto another target creature".
var testerOfTheTangentialMoveBody = game.ReflexiveBody("tester-of-the-tangential/move", testerOfTheTangentialMove,
	func(sourceID uuid.UUID, _ game.EffectParams) *game.TargetSpec {
		return TargetCreature("another target creature", OtherThan(sourceID))
	})

func testerOfTheTangentialPay(g *game.Game, item *game.StackItem) error {
	return MayPayX{
		Cost:  "{X}",
		Label: "Tester of the Tangential",
		Buys:  "move X +1/+1 counters onto another target creature",
		Goal:  testerOfTheTangentialGoal,
		Unit:  game.PayAmountCounters,
		Then:  testerOfTheTangentialPaid,
	}.Apply(NewContext(g, item))
}

// testerOfTheTangentialGoal is the X a bot pays: every +1/+1 counter on the
// Tester it can afford, when there is another creature of its own to take
// them.
func testerOfTheTangentialGoal(ctx *Context, ceiling int) int {
	self, ok := ctx.Game.LookupCardForEffect(ctx.Source())
	if !ok {
		return 0
	}
	for _, c := range ctx.Game.Battlefield.Cards {
		if c.InstanceID != self.InstanceID && c.Controller == ctx.Controller() && c.IsCreature() {
			return min(ceiling, self.Counters[game.CounterPlusOne])
		}
	}
	return 0
}

// testerOfTheTangentialMove removes up to X +1/+1 counters from the Tester
// and puts that many on the target. Nothing moves if the Tester is gone or a
// new object, or the target is.
func testerOfTheTangentialMove(g *game.Game, item *game.StackItem, p game.EffectParams) error {
	if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
		return nil
	}
	target := item.Targets[0].ID
	if _, ok := g.LookupCardForEffect(target); !ok || !onBattlefield(g, target) {
		return nil
	}
	if !onBattlefield(g, item.SourceCardID) || sourceIsNewObject(g, item) {
		return nil
	}
	ctx := NewContext(g, item)
	moved, err := socB9RemoveCounters(ctx, item.SourceCardID, game.CounterPlusOne, p.Amount)
	if err != nil || moved <= 0 {
		return err
	}
	return AddCounter{Target: target, Kind: game.CounterPlusOne, N: moved}.Apply(ctx)
}
