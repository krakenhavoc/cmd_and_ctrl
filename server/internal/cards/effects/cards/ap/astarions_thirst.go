package cardsap

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Astarion's Thirst — Instant {3}{B} (Commander Legends: Battle for Baldur's
// Gate, #5948):
//
//	"Exile target creature. Put X +1/+1 counters on a commander creature you
//	 control, where X is the power of the creature exiled this way."
//
// X is the target's power as it last was on the battlefield (CR 608.2h), read
// before it is exiled and carried to the exile's continuation, which runs only
// if the creature really reached exile (a replacement may send it elsewhere
// or a commander may be asked about the command zone, CR 903.9). A commander
// creature you control is chosen as the counters are put on: with one, no
// question; with several (partners), the controller names which (Amy Pond's
// option pick); with none, nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0236ad13-d8e4-49b4-b397-fe2f6136665d",
		Name:         "Astarion's Thirst",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok {
				return nil
			}
			c, ok := ctx.Game.LookupCardForEffect(t.ID)
			if !ok {
				return nil
			}
			ctx.Game.RecomputeLayersIfStaleLocked()
			return effects.ExileTarget{Target: t.ID, ThenTail: ctx.TailN(astarionsThirstExiled, max(c.CurrentPower(), 0))}.Apply(ctx)
		},
	})
}

// astarionsThirstExiled puts X counters on a commander creature you control,
// once the creature has reached exile. An on-disk key: never renamed, never
// reused.
var astarionsThirstExiled = effects.CardTail("astarions-thirst/exiled", func(ctx *effects.Context, o game.TailOutcome) error {
	x := o.Tail.N
	if len(o.IDs) == 0 || x <= 0 {
		return nil
	}
	var ids []uuid.UUID
	var options []game.ChoiceOption
	for _, c := range ctx.Game.BattlefieldCardsForEffect() {
		if c.Controller == ctx.Controller() && c.IsCommander && c.IsCreature() {
			ids = append(ids, c.InstanceID)
			options = append(options, game.ChoiceOption{Label: c.Name})
		}
	}
	switch len(ids) {
	case 0:
		return nil
	case 1:
		return effects.AddCounter{Target: ids[0], Kind: game.CounterPlusOne, N: x}.Apply(ctx)
	}
	return effects.PickOption{
		Question: "Astarion's Thirst — choose a commander creature you control to put the counters on",
		Options:  options,
		ThenKey:  astarionsThirstPicked,
		Carry:    ids,
		Nums:     []int{x},
	}.Apply(ctx)
})

var astarionsThirstPicked = effects.OptionPickThen("option-pick/astarions-thirst-commander", func(ctx *effects.Context, r game.OptionPicked) error {
	if r.Option == nil || r.Index < 0 || r.Index >= len(r.Carry) || len(r.Nums) == 0 {
		return nil
	}
	return effects.AddCounter{Target: r.Carry[r.Index], Kind: game.CounterPlusOne, N: r.Nums[0]}.Apply(ctx)
})
