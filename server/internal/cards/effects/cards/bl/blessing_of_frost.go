package cardsbl

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Blessing of Frost — Snow Sorcery {3}{G} (Kaldheim):
//
//	"Distribute X +1/+1 counters among any number of creatures you control,
//	 where X is the amount of {S} spent to cast this spell. Then draw a card
//	 for each creature you control with power 4 or greater."
//
// X is the snow mana that paid for the spell (CR 107.4h). The split is
// chosen as the spell resolves (blessingOfFrostResolve).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "240d650b-97ea-40ad-8c1b-fe1ea52d861f",
		Name:           "Blessing of Frost",
		Completeness:   effects.CompletenessFull,
		ReadsManaSpent: game.ReadsPoolManaOnly,
		WantsManaFrom:  game.ManaSourceSnow,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return blessingOfFrostResolve(ctx)
		},
	})
}

// --- Blessing of Frost ------------------------------------------------

// blessingOfFrostDrawForPower4 is "draw a card for each creature you control with
// power 4 or greater", counted from the board once the counters are on.
func blessingOfFrostDrawForPower4(ctx *effects.Context) error {
	n := 0
	for _, c := range ctx.Game.BattlefieldCardsForEffect() {
		if c.Controller == ctx.Controller() && c.IsCreature() && c.CurrentPower() >= 4 {
			n++
		}
	}
	return effects.DrawCards{Player: ctx.Controller(), N: n}.Apply(ctx)
}

// blessingOfFrostResolve is Blessing of Frost's whole effect: "Distribute X
// +1/+1 counters among any number of creatures you control, where X is the
// amount of {S} spent to cast this spell. Then draw a card for each creature
// you control with power 4 or greater."
//
// No creature is targeted, so the split is chosen as the effect resolves
// (CR 608.2d): the controller picks the creatures (each gets at least one
// counter), then names each share but the last, which takes what is left.
// The draw comes after the last share is placed, off the answer chain, so the
// counters count toward power 4. With X of 0, or no creature, nothing is
// distributed and the draw still happens.
func blessingOfFrostResolve(ctx *effects.Context) error {
	n := ctx.ManaSpent().CountFrom(game.ManaSourceSnow)
	var cands []uuid.UUID
	for _, c := range ctx.Game.BattlefieldCardsForEffect() {
		if c.Controller == ctx.Controller() && c.IsCreature() {
			cands = append(cands, c.InstanceID)
		}
	}
	if n <= 0 || len(cands) == 0 {
		return blessingOfFrostDrawForPower4(ctx)
	}
	if len(cands) == 1 {
		return blessingOfFrostShare(ctx, cands, n)
	}
	ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:   ctx.Controller(),
		Source:    ctx.Source(),
		Question:  "Blessing of Frost — choose any number of creatures you control to distribute +1/+1 counters among (each gets at least one)",
		Cards:     cands,
		Min:       1,
		Max:       min(n, len(cands)),
		ThenKey:   blessingOfFrostPick,
		CarryNums: []int{n},
		LiveZone:  game.ZoneBattlefield,
		Asked:     game.NewAskedItem(ctx.Item),
	})
	return nil
}

var blessingOfFrostPick = effects.AskedPick("choose-cards/s82-old-b061-blessing-of-frost-creatures", func(ctx *effects.Context, r game.CardsChosen) error {
	if len(r.Nums) == 0 {
		return nil
	}
	var picked []uuid.UUID
	for _, id := range r.Picked {
		if c, ok := ctx.Game.LookupCardForEffect(id); ok && c.Controller == ctx.Controller() && c.IsCreature() {
			picked = append(picked, id)
		}
	}
	return blessingOfFrostShare(ctx, picked, r.Nums[0])
})

// blessingOfFrostShare gives the first creature its share: the last takes what is
// left, any other is asked, from one up to what leaves one for each of the
// rest. When nothing is left to give it draws.
func blessingOfFrostShare(ctx *effects.Context, creatures []uuid.UUID, remaining int) error {
	switch {
	case len(creatures) == 0 || remaining <= 0:
		return blessingOfFrostDrawForPower4(ctx)
	case len(creatures) == 1:
		if z := ctx.Game.FindCardZoneForEffect(creatures[0]); z != nil && z.Kind == game.ZoneBattlefield {
			if err := (effects.AddCounter{Target: creatures[0], Kind: game.CounterPlusOne, N: remaining}).Apply(ctx); err != nil {
				return err
			}
		}
		return blessingOfFrostDrawForPower4(ctx)
	}
	return effects.ChooseNumber{
		Min:      1,
		Max:      remaining - (len(creatures) - 1),
		Question: "Blessing of Frost — how many +1/+1 counters go on the next creature? (each remaining creature needs at least one)",
		Unit:     game.PayAmountOther,
		Then:     blessingOfFrostShareThen,
		Carry:    creatures,
		Nums:     []int{remaining},
	}.Apply(ctx)
}

var blessingOfFrostShareThen game.NumberThen

// Registered from init: the answer asks the next share through
// blessingOfFrostShare, which names it, so a var initialiser would be a cycle.
func init() {
	blessingOfFrostShareThen = effects.NumberThen("number/s82-old-b061-blessing-of-frost-share", func(ctx *effects.Context, r game.NumberAnswer) error {
		if len(r.Carry) == 0 || len(r.Nums) == 0 {
			return nil
		}
		share := min(max(r.Number, 1), r.Nums[0]-(len(r.Carry)-1))
		if z := ctx.Game.FindCardZoneForEffect(r.Carry[0]); z != nil && z.Kind == game.ZoneBattlefield {
			if err := (effects.AddCounter{Target: r.Carry[0], Kind: game.CounterPlusOne, N: share}).Apply(ctx); err != nil {
				return err
			}
		}
		return blessingOfFrostShare(ctx, r.Carry[1:], r.Nums[0]-share)
	})
}
