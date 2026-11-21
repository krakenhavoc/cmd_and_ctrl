package cardslu

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func init() {
	masterWildHuntChosen = effects.NumberThen("number/master-of-the-wild-hunt-divide-return-damage", func(ctx *effects.Context, r game.NumberAnswer) error {
		if len(r.Carry) < 2 || len(r.Nums) < 2 {
			return nil
		}
		shares := append(append([]int(nil), r.Nums[2:]...), r.Number)
		return masterWildHuntAsk(ctx, r.Chooser, r.Carry, r.Nums[1], shares, r.Nums[0]-r.Number)
	})

	effects.Register(effects.Spec{OracleID: "dd25cab9-15cc-40f5-9b68-8f345dd7e952", Name: "Master of the Wild Hunt", Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{effects.AtYourUpkeep("Master of the Wild Hunt — create a Wolf", effects.Do(effects.CreateToken{Template: effects.TokenCard("2/2 green Wolf"), N: 1}))},
		Activated: []effects.ActivatedAbility{{Label: "{T}: Tap all untapped Wolf creatures you control. Each Wolf tapped this way deals damage equal to its power to target creature. That creature deals damage equal to its power divided as its controller chooses among any number of those Wolves.", Cost: effects.TapCost(), Targets: effects.TargetCreature("target creature"), Effect: masterWildHunt}},
	})
	masterWildHuntDamage = effects.RegisterThenSteps("then-steps/master-of-the-wild-hunt-damage",
		effects.Instruction(masterWildHuntWolvesDeal), effects.Instruction(masterWildHuntDivide))
}

const masterWildHuntRecord = "master-of-the-wild-hunt/wolves"

var masterWildHuntDamage effects.ThenSteps

func masterWildHunt(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	target, ok := effects.B16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	// Tapping is an effect on resolution. Already tapped Wolves and Wolf
	// noncreatures neither participate nor receive the return damage.
	ids := []uuid.UUID{target}
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == ctx.Controller() && c.IsCreature() && c.HasSubtype(ctx.TextSubtype("Wolf")) && !c.Tapped {
			if err := g.TapTargetForEffect(c.InstanceID); err != nil {
				return err
			}
			ids = append(ids, c.InstanceID)
		}
	}
	if len(ids) == 1 {
		return nil
	}
	epochs := make([]int, len(ids))
	for i, id := range ids {
		ref, _ := g.PermanentRefForEffect(id)
		epochs[i] = ref.Epoch
	}
	ctx.RecordStep(masterWildHuntRecord, ids, epochs...)
	return masterWildHuntDamage.Run(ctx)
}

func masterWildHuntWolvesDeal(ctx *effects.Context) error {
	r := ctx.StepRecord(masterWildHuntRecord)
	ids := r.IDs
	if len(ids) < 2 {
		return nil
	}
	// The Wolves deal damage together. The creature's return damage is a
	// later instruction, after all replacements for this damage settle.
	powers := make([]int, len(ids)-1)
	for i, id := range ids[1:] {
		info, _ := ctx.Game.PermanentForEffect(game.ObjectRef{ID: id, Epoch: r.Ns[i+1]})
		powers[i] = info.Power
	}
	return ctx.Game.DamageInstanceForEffect(func() error {
		for i, id := range ids[1:] {
			ref := game.ObjectRef{ID: id, Epoch: r.Ns[i+1]}
			if err := (effects.DealDamage{SourceObject: &ref, Target: ids[0], Amount: powers[i]}).Apply(ctx); err != nil {
				return err
			}
		}
		return nil
	})
}

func masterWildHuntDivide(ctx *effects.Context) error {
	r := ctx.StepRecord(masterWildHuntRecord)
	ids := r.IDs
	if len(ids) < 2 {
		return nil
	}
	c, ok := ctx.Game.PermanentForEffect(game.ObjectRef{ID: ids[0], Epoch: r.Ns[0]})
	if !ok {
		return nil
	}
	return masterWildHuntAsk(ctx, c.Controller, ids, r.Ns[0], nil, max(0, c.Power))
}

// Carry is the creature followed by the Wolves. Nums is the remaining
// damage, the source object's epoch, then completed shares. Zero shares allow any subset of
// the Wolves; these are choices, never additional targets.
func masterWildHuntAsk(ctx *effects.Context, chooser uuid.UUID, ids []uuid.UUID, epoch int, shares []int, remaining int) error {
	if len(shares) == len(ids)-2 || remaining == 0 {
		for len(shares) < len(ids)-2 {
			shares = append(shares, 0)
		}
		shares = append(shares, remaining)
		return ctx.Game.DamageInstanceForEffect(func() error {
			for i, id := range ids[1:] {
				ref := game.ObjectRef{ID: ids[0], Epoch: epoch}
				if err := (effects.DealDamage{SourceObject: &ref, Target: id, Amount: shares[i]}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		})
	}
	wolf, _ := ctx.Game.LookupCardForEffect(ids[len(shares)+1])
	return (effects.ChooseNumber{Chooser: chooser, Min: 0, Max: remaining, Unit: game.PayAmountDamage,
		Question: fmt.Sprintf("Master of the Wild Hunt — damage to %s (%d left to divide)", wolf.Name, remaining),
		Then:     masterWildHuntChosen, Carry: ids, Nums: append([]int{remaining, epoch}, shares...),
	}).Apply(ctx)
}

var masterWildHuntChosen game.NumberThen
