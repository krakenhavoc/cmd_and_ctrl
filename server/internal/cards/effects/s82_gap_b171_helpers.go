package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Helpers for S82 GAP batch 171 (Jumpstart 2022 instants, #5948).

// GapB171Wolves is Predator's Howl's body: one 2/2 green Wolf, or three if a
// creature died this turn (morbid, any creature, read as it resolves).
func GapB171Wolves(item *game.StackItem, ctx *Context) error {
	n := 1
	if B11CreaturesDiedThisTurn(ctx.Game) > 0 {
		n = 3
	}
	return CreateToken{Controller: item.Controller, Template: TokenCard("2/2 green Wolf"), N: n}.Apply(ctx)
}

// gapB171AnimateFetchedLand is Rampaging Growth's second sentence: "Until end
// of turn, that land becomes a 4/3 Insect creature with reach and haste. It's
// still a land." It runs as the search's tail, told the land that was found
// (none when the search found nothing), and pins the effect to that object
// (CR 611.2c). An on-disk key.
var gapB171AnimateFetchedLand = CardTail("rampaging-growth/animate", func(ctx *Context, o game.TailOutcome) error {
	for _, id := range o.IDs {
		if !OnBattlefield(ctx.Game, id) {
			continue
		}
		if err := Cmdr2B13Animate(ctx, id, "Rampaging Growth — 4/3 Insect creature with reach and haste until end of turn", 4, 3, "Insect", nil, "reach", "haste"); err != nil {
			return err
		}
	}
	return nil
})

// GapB171RampagingGrowth searches for a basic land card, puts it onto the
// battlefield (untapped) and shuffles, then animates that land until end of
// turn.
func GapB171RampagingGrowth(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	return SearchLibrary{
		Player:    item.Controller,
		Predicate: IsBasicLand,
		Dest:      game.ZoneBattlefield,
		Limit:     1,
		Shuffle:   true,
		Reason:    "Rampaging Growth — a basic land card to put onto the battlefield",
		Source:    item.SourceCardID,
		ThenTail:  ctx.Tail(gapB171AnimateFetchedLand),
	}.Apply(ctx)
}

// GapB171DamageFirstTarget is "deals N damage to target creature" as an
// Effect, for a trigger: the first target, if it is still legal as the
// ability resolves (CR 608.2b).
func GapB171DamageFirstTarget(n int) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		if len(item.Targets) == 0 || !ctx.IsTargetLegal(item.Targets[0]) {
			return nil
		}
		return DealDamage{Source: ctx.Source(), Target: item.Targets[0].ID, Amount: n}.Apply(ctx)
	}
}

// GapB171Outnumber is "deals damage to target creature equal to the number of
// creatures you control", counted as the spell resolves through the layers.
func GapB171Outnumber(item *game.StackItem, ctx *Context) error {
	id, ok := B16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	n := CountControlled(ctx.Game, item.Controller, func(c game.Card) bool { return c.IsCreature() })
	if n <= 0 {
		return nil
	}
	return DealDamage{Source: ctx.Source(), Target: id, Amount: n}.Apply(ctx)
}
