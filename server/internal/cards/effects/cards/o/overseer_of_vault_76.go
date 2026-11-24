package cardso

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Overseer of Vault 76 — Legendary Creature — Human Advisor {2}{W}, 3/3
// (Fallout, #6100):
//
//	"First Contact — Whenever Overseer of Vault 76 or another creature you
//	 control with power 3 or less enters, put a quest counter on Overseer
//	 of Vault 76.
//	 At the beginning of combat on your turn, you may remove three quest
//	 counters from among permanents you control. When you do, put a +1/+1
//	 counter on each creature you control and they gain vigilance until
//	 end of turn."
//
// First Contact reads the entering creature's power as it is on the
// battlefield; the Overseer itself always counts.
//
// The combat trigger asks, as it resolves, whether to remove three quest
// counters, and only when the permanents you control hold three in all
// (CR 608.2c: an action that can't be done fully isn't offered). On "yes"
// you split the three among them: one question per permanent with quest
// counters, in battlefield order, each bounded by what it has and by what
// the rest can still cover, and asked only when there is a choice. Once
// three really came off, the CR 603.12 reflexive trigger goes on the
// stack: a +1/+1 counter on each creature you control as it resolves, and
// those creatures gain vigilance until end of turn (the set is fixed then,
// CR 611.2c).
//
// No simplification.
const overseerName = "Overseer of Vault 76"

var (
	// overseerAsk is the "you may" answer. An on-disk key: never renamed,
	// never reused.
	overseerAsk = effects.MayYes("confirm/overseer-of-vault-76-remove-three-quest-counters", func(ctx *effects.Context, _ game.ConfirmAnswer) error {
		perms, have := overseerQuestHolders(ctx.Game, ctx.Controller())
		return overseerSplit(ctx, perms, have, make([]int, len(perms)), 0)
	})

	// overseerBoostBody is the reflexive trigger. An on-disk key: never
	// renamed, never reused.
	overseerBoostBody = game.ReflexiveBody("overseer-of-vault-76/counters-and-vigilance", effects.SimpleBody(overseerBoost), nil)
)

// overseerSplitAnswer is one permanent's share of the three. Assigned in
// init, since its body asks the next share through it. An on-disk key:
// never renamed, never reused.
var overseerSplitAnswer game.NumberThen

func init() {
	overseerSplitAnswer = effects.NumberThen("number/overseer-of-vault-76-split-quest-counters", overseerShare)
	effects.Register(effects.Spec{
		OracleID:     "84f32a1d-3861-418e-a3c2-7e731a2cd55c",
		Name:         overseerName,
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, overseerFirstContact,
				overseerName+" — First Contact — put a quest counter on Overseer of Vault 76",
				effects.Do(effects.CounterOnThis{Kind: "quest", N: 1})),
			effects.AtBeginningOfYourCombat(overseerName+" — you may remove three quest counters from among permanents you control", overseerOffer),
		},
	})
}

// overseerShare records one permanent's share and asks the next.
func overseerShare(ctx *effects.Context, r game.NumberAnswer) error {
	perms := r.Carry
	n := len(perms)
	if len(r.Nums) != 1+2*n {
		return nil
	}
	idx := r.Nums[0]
	have := append([]int(nil), r.Nums[1:1+n]...)
	picked := append([]int(nil), r.Nums[1+n:]...)
	if idx < 0 || idx >= n {
		return nil
	}
	picked[idx] = r.Number
	return overseerSplit(ctx, perms, have, picked, idx+1)
}

// overseerFirstContact is "Overseer of Vault 76 or another creature you
// control with power 3 or less enters".
func overseerFirstContact(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind == game.EventETB && ev.CardID == source.InstanceID {
		return true
	}
	c, ok := effects.EnteredUnderYourControl(ev, source, g, true)
	return ok && c.IsCreature() && c.CurrentPower() <= 3
}

// overseerQuestHolders is the permanents `you` control with quest counters,
// in battlefield order, and how many each has.
func overseerQuestHolders(g *game.Game, you uuid.UUID) ([]uuid.UUID, []int) {
	var perms []uuid.UUID
	var have []int
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == you && c.Counters["quest"] > 0 {
			perms = append(perms, c.InstanceID)
			have = append(have, c.Counters["quest"])
		}
	}
	return perms, have
}

// overseerOffer asks the "you may" when three quest counters are there.
func overseerOffer(g *game.Game, item *game.StackItem) error {
	_, have := overseerQuestHolders(g, item.Controller)
	total := 0
	for _, n := range have {
		total += n
	}
	if total < 3 {
		return nil
	}
	return effects.MayChoice{
		Question: overseerName + " — remove three quest counters from among permanents you control?",
		Then:     overseerAsk,
	}.Apply(effects.NewContext(g, item))
}

// overseerSplit asks for permanent idx's share, or, once every share is
// known, removes them.
func overseerSplit(ctx *effects.Context, perms []uuid.UUID, have, picked []int, idx int) error {
	for ; idx < len(perms); idx++ {
		left := 3
		for _, p := range picked[:idx] {
			left -= p
		}
		later := 0
		for _, n := range have[idx+1:] {
			later += n
		}
		most := min(have[idx], left)
		least := max(0, left-later)
		if least >= most {
			picked[idx] = most
			continue
		}
		nums := append(append([]int{idx}, have...), picked...)
		return effects.ChooseNumber{
			Min:      least,
			Max:      most,
			Question: overseerName + " — quest counters to remove from this permanent",
			Unit:     game.PayAmountCounters,
			Then:     overseerSplitAnswer,
			Carry:    append([]uuid.UUID(nil), perms...),
			Nums:     nums,
		}.Apply(ctx)
	}
	removed := 0
	for i, id := range perms {
		if picked[i] <= 0 {
			continue
		}
		before := 0
		if c, ok := ctx.Game.LookupCardForEffect(id); ok && c.Controller == ctx.Controller() {
			before = c.Counters["quest"]
		} else {
			continue
		}
		if err := (effects.AddCounter{Target: id, Kind: "quest", N: -picked[i]}).Apply(ctx); err != nil {
			return err
		}
		if c, ok := ctx.Game.LookupCardForEffect(id); ok {
			removed += before - c.Counters["quest"]
		}
	}
	if removed < 3 {
		return nil
	}
	return effects.WhenYouDo(overseerName+" — put a +1/+1 counter on each creature you control; they gain vigilance until end of turn", overseerBoostBody).Apply(ctx)
}

// overseerBoost is the reflexive trigger's resolution.
func overseerBoost(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	var mine []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == item.Controller && c.IsCreature() {
			mine = append(mine, c.InstanceID)
		}
	}
	for _, id := range mine {
		if err := (effects.AddCounter{Target: id, Kind: game.CounterPlusOne, N: 1}).Apply(ctx.AsGroupMember()); err != nil {
			return err
		}
	}
	for _, id := range mine {
		if err := (effects.GrantKeywordUntilEOT{
			Target:   id,
			Keywords: []string{"vigilance"},
			Label:    overseerName + " — vigilance until end of turn",
		}).Apply(ctx.AsGroupMember()); err != nil {
			return err
		}
	}
	return nil
}
