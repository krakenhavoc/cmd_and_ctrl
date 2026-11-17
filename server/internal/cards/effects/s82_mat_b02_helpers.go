package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_mat_b02_helpers.go — shared pieces of the March of the Machine: The
// Aftermath batch 2 cards (#5476, S82).

// matArtifactOrEnchantmentFromYourGraveyard is the target clause of
// "return target artifact or enchantment card from your graveyard to the
// battlefield" (Harnessed Snubhorn, Campus Renovation).
func matArtifactOrEnchantmentFromYourGraveyard(text string) *game.TargetSpec {
	return TargetCardInGraveyard(text, Or(Artifact(), Enchantment()), YouOwn())
}

// matReturnTargetToBattlefield returns the card the ability or spell
// targeted to the battlefield under its controller's control. A target
// that left the graveyard in response (or no target at all, for "up to
// one") does nothing (CR 608.2b).
func matReturnTargetToBattlefield(ctx *Context) {
	_, _ = reanimateSingleTarget(ctx, ctx.Controller())
}

// matDivideCreatureCardsInYourGraveyard is "distribute X counters …, where X
// is the number of creature cards in your graveyard as you cast this spell"
// (Undercity Upheaval): read once as the targets are announced (CR 601.2d),
// so a creature card that leaves the graveyard in response does not change
// how many counters are divided.
var matDivideCreatureCardsInYourGraveyard = game.RegisterDivideAmount("mat-creature-cards-in-your-graveyard",
	func(g *game.Game, a game.DivideAmountArgs) game.DivideSpec {
		return game.DivideSpec{Total: b11CreatureCardsInGraveyard(g, a.Controller)}
	})

// matControlledCreatures are the creatures `player` controls, through the
// layered view, in battlefield order.
func matControlledCreatures(g *game.Game, player uuid.UUID) []uuid.UUID {
	var out []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == player && c.IsCreature() {
			out = append(out, c.InstanceID)
		}
	}
	return out
}

// matDistributePlusOneCounters is "distribute N +1/+1 counters among any
// number of creatures you control" with no target word (Feast of the
// Victorious Dead): the split is chosen as the effect resolves (CR 608.2d)
// and not announced with targets. The controller picks one to N of their
// creatures (each chosen creature gets at least one counter, as in CR
// 601.2d's "distribute"), then names each share but the last, which takes
// what is left. One creature takes all N; no creature, or N of 0, does
// nothing. A creature that left the battlefield between the choices loses
// its share.
func matDistributePlusOneCounters(ctx *Context, n int) error {
	cands := matControlledCreatures(ctx.Game, ctx.Controller())
	if n <= 0 || len(cands) == 0 {
		return nil
	}
	if len(cands) == 1 {
		return matPutCounters(ctx, cands[0], n)
	}
	ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:   ctx.Controller(),
		Source:    ctx.Source(),
		Question:  "Choose any number of creatures you control to distribute +1/+1 counters among (each gets at least one)",
		Cards:     cands,
		Min:       1,
		Max:       min(n, len(cands)),
		ThenKey:   matDistributePick,
		CarryNums: []int{n},
		LiveZone:  game.ZoneBattlefield,
		Asked:     game.NewAskedItem(ctx.Item),
	})
	return nil
}

// matDistributePick is the creatures chosen: it starts the shares.
var matDistributePick = askedPick("choose-cards/mat-distribute-counters-among", func(ctx *Context, r game.CardsChosen) error {
	if len(r.Nums) == 0 {
		return nil
	}
	var picked []uuid.UUID
	for _, id := range r.Picked {
		if c, ok := ctx.Game.LookupCardForEffect(id); ok && c.Controller == ctx.Controller() && c.IsCreature() {
			picked = append(picked, id)
		}
	}
	return matDistributeStep(ctx, picked, r.Nums[0])
})

// matDistributeStep gives the first creature its share: the last takes what
// is left, any other is asked, from one up to what leaves one for each of
// the rest.
func matDistributeStep(ctx *Context, creatures []uuid.UUID, remaining int) error {
	switch {
	case len(creatures) == 0 || remaining <= 0:
		return nil
	case len(creatures) == 1:
		return matPutCounters(ctx, creatures[0], remaining)
	}
	return ChooseNumber{
		Min:      1,
		Max:      remaining - (len(creatures) - 1),
		Question: "How many +1/+1 counters go on the next creature? (each remaining creature needs at least one)",
		Unit:     game.PayAmountOther,
		Then:     matDistributeShare,
		Carry:    creatures,
		Nums:     []int{remaining},
	}.Apply(ctx)
}

// matDistributeShare puts the chosen share on the first carried creature and
// goes on to the rest.
var matDistributeShare game.NumberThen

// The branch is registered from init: it asks the next share through
// matDistributeStep, which names it, so a var initialiser would be a cycle.
func init() {
	matDistributeShare = NumberThen("number/mat-distribute-counters-share", matDistributeShareBody)
}

func matDistributeShareBody(ctx *Context, r game.NumberAnswer) error {
	if len(r.Carry) == 0 || len(r.Nums) == 0 {
		return nil
	}
	share := min(max(r.Number, 1), r.Nums[0]-(len(r.Carry)-1))
	if err := matPutCounters(ctx, r.Carry[0], share); err != nil {
		return err
	}
	return matDistributeStep(ctx, r.Carry[1:], r.Nums[0]-share)
}

// matPutCounters puts n +1/+1 counters on a creature that is still on the
// battlefield.
func matPutCounters(ctx *Context, id uuid.UUID, n int) error {
	if z := ctx.Game.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneBattlefield {
		return nil
	}
	return AddCounter{Target: id, Kind: game.CounterPlusOne, N: n}.Apply(ctx)
}
