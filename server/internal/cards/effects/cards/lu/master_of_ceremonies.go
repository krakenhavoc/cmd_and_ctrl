package cardslu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Master of Ceremonies — Creature — Rhino Druid {3}{W}, 3/4 (Streets of New
// Capenna, #5491):
//
//	"At the beginning of your upkeep, each opponent chooses money, friends, or
//	 secrets. For each player who chose money, you and that player each create a
//	 Treasure token. For each player who chose friends, you and that player each
//	 create a 1/1 green and white Citizen creature token. For each player who
//	 chose secrets, you and that player each draw a card."
//
// Each opponent is asked in seat order (an option-pick prompt addressed to
// them); once every answer is in, the three clauses run in printed order, money
// then friends then secrets, and within one clause the players who chose it in
// the order they were asked, you first and then that player. The work is a keyed
// step list over a step record (the players and their clause), so a table
// waiting on a draw's prompt, or on an answer, is a restore point. An opponent
// who has left the game before answering is skipped. Whatever the controller
// does with the effect is the ability's controller's, so it still resolves in
// full if the Rhino has left the battlefield.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "18ed0c8a-db8f-4247-87c4-544b833f01bd",
		Name:         "Master of Ceremonies",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Master of Ceremonies — each opponent chooses money, friends, or secrets",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return masterOfCeremoniesAsk(ctx, ctx.Opponents(), nil, nil)
				}),
		},
	})
	masterOfCeremoniesAnswered = effects.OptionPickThen("option-pick/master-of-ceremonies-answer", func(ctx *effects.Context, r game.OptionPicked) error {
		if len(r.Nums) < 1 || r.Nums[0] < 0 || r.Nums[0] > len(r.Carry) || r.Nums[0] > len(r.Nums)-1 {
			return nil
		}
		n := r.Nums[0]
		answered := append([]uuid.UUID(nil), r.Carry[:n]...)
		kinds := append([]int(nil), r.Nums[1:1+n]...)
		rest := r.Carry[n:]
		if r.Index >= 0 && r.Index < len(masterOfCeremoniesOptions) {
			answered = append(answered, r.Chooser)
			kinds = append(kinds, r.Index)
		}
		return masterOfCeremoniesAsk(ctx, rest, answered, kinds)
	})
	masterOfCeremoniesSteps = effects.RegisterThenSteps("then-steps/master-of-ceremonies-clauses",
		effects.Instruction(func(ctx *effects.Context) error { return masterOfCeremoniesAct(ctx, ctx.Controller()) }),
		effects.Instruction(func(ctx *effects.Context) error {
			rec := ctx.StepRecord(masterOfCeremoniesRecord)
			if len(rec.IDs) == 0 {
				return nil
			}
			return masterOfCeremoniesAct(ctx, rec.IDs[0])
		}),
		effects.Instruction(func(ctx *effects.Context) error {
			rec := ctx.StepRecord(masterOfCeremoniesRecord)
			if len(rec.IDs) <= 1 {
				return nil
			}
			ctx.RecordStep(masterOfCeremoniesRecord, rec.IDs[1:], rec.Ns[1:]...)
			return masterOfCeremoniesSteps.Run(ctx)
		}),
	)
}

// masterOfCeremoniesOptions are the three words, in the printed order.
var masterOfCeremoniesOptions = []game.ChoiceOption{
	{Label: "Money — the caster and you each create a Treasure token"},
	{Label: "Friends — the caster and you each create a 1/1 green and white Citizen creature token"},
	{Label: "Secrets — the caster and you each draw a card"},
}

// masterOfCeremoniesRecord files the players still owed their clause, with the
// clause each chose, front first.
const masterOfCeremoniesRecord = "master-of-ceremonies/owed"

// masterOfCeremoniesAnswered is one opponent's answer, keyed (#4253): the seat
// that answered is the chooser; the seats already answered then the seats still
// to ask ride the carry, and the number answered then their clauses ride the
// numbers. An on-disk key: never renamed, never reused. Registered in init
// because the walk asks again from inside it.
var masterOfCeremoniesAnswered game.OptionPickThen

// masterOfCeremoniesSteps is registered in init, because its last step runs it
// again for the next player owed.
var masterOfCeremoniesSteps effects.ThenSteps

func masterOfCeremoniesAsk(ctx *effects.Context, remaining, answered []uuid.UUID, kinds []int) error {
	for len(remaining) > 0 {
		opp := remaining[0]
		rest := remaining[1:]
		if p := ctx.Game.PlayerByIDForEffect(opp); p == nil || p.Eliminated {
			remaining = rest
			continue
		}
		return effects.PickOption{
			Player:   opp,
			Question: "Master of Ceremonies — choose money, friends, or secrets",
			Options:  masterOfCeremoniesOptions,
			ThenKey:  masterOfCeremoniesAnswered,
			WithItem: true,
			Carry:    append(append([]uuid.UUID(nil), answered...), rest...),
			Nums:     append([]int{len(answered)}, kinds...),
		}.Apply(ctx)
	}
	// Every answer is in: money, then friends, then secrets, each in the order
	// the players were asked.
	var owed []uuid.UUID
	var clauses []int
	for kind := range masterOfCeremoniesOptions {
		for i, id := range answered {
			if kinds[i] == kind {
				owed = append(owed, id)
				clauses = append(clauses, kind)
			}
		}
	}
	if len(owed) == 0 {
		return nil
	}
	ctx.RecordStep(masterOfCeremoniesRecord, owed, clauses...)
	return masterOfCeremoniesSteps.Run(ctx)
}

// masterOfCeremoniesAct does the front clause of the record for `who`.
func masterOfCeremoniesAct(ctx *effects.Context, who uuid.UUID) error {
	rec := ctx.StepRecord(masterOfCeremoniesRecord)
	if len(rec.Ns) == 0 {
		return nil
	}
	switch rec.Ns[0] {
	case 0:
		return effects.CreateToken{Controller: who, Template: effects.TreasureToken(), N: 1}.Apply(ctx)
	case 1:
		return effects.CreateToken{Controller: who, Template: effects.TokenCard("1/1 green and white Citizen"), N: 1}.Apply(ctx)
	default:
		return effects.DrawCards{Player: who, N: 1}.Apply(ctx)
	}
}
