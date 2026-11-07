package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Aang, Master of Elements — Legendary Creature — Avatar Ally 6/6, the
// back face of Avatar Aang (avatar_aang.go; #4459, seam #5276):
//
//	"Flying
//	 Spells you cast cost {W}{U}{B}{R}{G} less to cast. (This can reduce
//	 generic costs.)
//	 At the beginning of each upkeep, you may transform Aang, Master of
//	 Elements. If you do, you gain 4 life, draw four cards, put four
//	 +1/+1 counters on him, and he deals 4 damage to each opponent."
//
// The discount is a cast reduction with a mana unit of {W}{U}{B}{R}{G},
// spent by CR 118.7: each coloured symbol takes one mana of its colour off
// the cost, or one generic mana when the cost has none of that colour left
// (CR 118.7b, the reminder text), and nothing goes below {0} (CR 601.2f).
//
// The upkeep trigger is every player's upkeep. Its "you may" is asked as
// it resolves (CR 603.5). "Transform" is the permanent turning over in
// place, only if it hasn't transformed since the trigger was put on the
// stack (CR 701.27f); "if you do" is that it did. Then the four
// instructions run in order, the later ones waiting for any prompt the
// draw asks (CR 608.2c). "Him" is the same permanent, now Avatar Aang
// (CR 712.18), so the counters land on him and the damage is his.
//
// No simplification.
func init() {
	reduction := CostsLess(1, "Spells you cast cost {W}{U}{B}{R}{G} less to cast.", YourSpell())
	unit, err := game.ParseCost("{W}{U}{B}{R}{G}")
	if err != nil {
		panic(err)
	}
	reduction.Unit = &unit
	Register(Spec{
		OracleID:        avatarAangOracle + "#1",
		Name:            "Aang, Master of Elements",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		CostModifiers:   []game.CostModifier{reduction},
		Triggered: []game.TriggeredAbility{
			Optional(On(game.EventBeginUpkeep, AnyPlayer,
				"Aang, Master of Elements — you may transform him; if you do, gain 4 life, draw four cards, four +1/+1 counters on him, 4 damage to each opponent",
				Do(
					Instruction(aangMasterTransform),
					If{Cond: aangMasterTransformed{}, Then: []Applier{
						GainLife{Amount: 4},
						DrawCards{N: 4},
						CounterOnThis{Kind: game.CounterPlusOne, N: 4},
						DamageEachOpponent{N: 4},
					}},
				)),
				"Aang, Master of Elements — transform him? (Gain 4 life, draw four cards, four +1/+1 counters, 4 damage to each opponent.)"),
		},
	})
}

// aangMasterTransform is "you may transform Aang, Master of Elements",
// answered yes: it turns him over if he hasn't turned over since the
// trigger was put on the stack (CR 701.27f) and is still on the
// battlefield, and marks the item when he did, for "if you do".
func aangMasterTransform(ctx *Context) error {
	if sourceTurnedOverSince(ctx) || !onBattlefield(ctx.Game, ctx.Source()) {
		return nil
	}
	if err := (TransformThis{}).Apply(ctx); err != nil {
		return err
	}
	if sourceTurnedOverSince(ctx) {
		ctx.Item.Params.Amount = aangMasterDidTransform
	}
	return nil
}

// aangMasterDidTransform marks the resolving item's Params.Amount once
// its transform happened. Nothing else on this row reads Amount.
const aangMasterDidTransform = 1

// aangMasterTransformed is "if you do".
type aangMasterTransformed struct{}

func (aangMasterTransformed) Holds(ctx *Context) bool {
	return ctx.Item != nil && ctx.Item.Params.Amount == aangMasterDidTransform
}

func (aangMasterTransformed) condition() (game.FootprintStep, bool) {
	return game.FootprintStep{}, false
}
