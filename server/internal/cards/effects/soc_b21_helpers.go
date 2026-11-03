package effects

import (
	"fmt"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// soc_b21_helpers.go — helpers for Secrets of Strixhaven batch 21 (#4456).

// seethingSongResolve is "Add {R}{R}{R}{R}{R}." as a spell's whole body:
// Seething Song, and the prepare spell of Blazing Firesinger.
func seethingSongResolve(item *game.StackItem, ctx *Context) error {
	return AddMana{Player: item.Controller, Produced: "{R}{R}{R}{R}{R}"}.Apply(ctx)
}

// socB21CostsLessManaIfItTargets is "This spell costs {1}{U} less to cast if
// it targets [a card that matches]" (Brush Off): CostsLessIfItTargets for a
// reduction that is not plain generic mana. `unit` is the mana taken off,
// spent by CR 118.7 (a coloured symbol removes one mana of its colour, or one
// generic when the cost has none of it left; the cost never goes below {0},
// CR 601.2f). Panics at init on a string the parser rejects.
func socB21CostsLessManaIfItTargets(unit, label string, match CardPredicate) game.CostModifier {
	u, err := game.ParseCost(unit)
	if err != nil || u.XSlots > 0 || u.ManaValue() == 0 {
		panic(fmt.Sprintf("effects.socB21CostsLessManaIfItTargets: %q is not a fixed mana amount (%v)", unit, err))
	}
	m := CostsLessIfItTargets(1, label, match)
	m.Unit = &u
	return m
}

// socB21CastAnotherInstantOrSorcery is "if you've cast another instant or
// sorcery spell this turn", read by the spell that is resolving. The tally
// counts every instant and sorcery spell the controller cast this turn,
// this one included when it was cast; a copy was never cast, so the spell it
// copies counts as the other one.
func socB21CastAnotherInstantOrSorcery(item *game.StackItem, ctx *Context) bool {
	need := 2
	if item.IsCopy {
		need = 1
	}
	return ctx.Game.CastTallyFor(ctx.Controller()).InstantSorcery >= need
}

// socB21BurrogBarrage is Burrog Barrage's whole resolution, read off a
// two-clause announcement: slot 0 is the creature you control, slot 1 the
// "up to one" opposing creature. The pump comes first and stays when the
// damage cannot happen; a gone creature of yours deals nothing.
func socB21BurrogBarrage(item *game.StackItem, ctx *Context) error {
	mine, ok := ctx.ClauseTarget(0)
	if !ok || mine.Kind != game.TargetCard {
		return nil
	}
	if socB21CastAnotherInstantOrSorcery(item, ctx) {
		if err := (BoostUntilEOT{Target: mine.ID, Power: 1, Label: "Burrog Barrage — +1/+0"}).Apply(ctx); err != nil {
			return err
		}
	}
	theirs, ok := ctx.ClauseTarget(1)
	if !ok || theirs.Kind != game.TargetCard {
		return nil
	}
	// The layers are brought current first: the +1/+0 above is a scoped
	// effect the cached view has not seen yet.
	return DealDamage{Source: mine.ID, Target: theirs.ID, Amount: b31CurrentPowerOf(ctx.Game, mine.ID)}.Apply(ctx)
}

// borrowedKnowledgeOpponentsHand is the draw of the first bullet, run once
// the discard has settled: a card for each card in the target opponent's
// hand, counted then (CR 608.2c). An on-disk key.
var borrowedKnowledgeOpponentsHand = CardTail("borrowed-knowledge/draw-opponents-hand", func(ctx *Context, _ game.TailOutcome) error {
	t, ok := ModeTarget(ctx, 0)
	if !ok || t.Kind != game.TargetPlayer {
		return nil
	}
	p := ctx.Game.PlayerByIDForEffect(t.ID)
	if p == nil {
		return nil
	}
	return DrawCards{Player: ctx.Controller(), N: p.Hand.Size()}.Apply(ctx)
})

// socB21DiscardHandThen discards the controller's whole hand and runs `then`
// once the discard has settled, told the cards that really reached a graveyard
// (Forgotten Creation's shape). It runs even for an empty hand.
func socB21DiscardHandThen(ctx *Context, item *game.StackItem, then game.TailRef) error {
	p := ctx.Game.PlayerByIDForEffect(item.Controller)
	if p == nil {
		return nil
	}
	return ctx.Game.DiscardRandomKeyedForEffect(item.Controller, p.Hand.Size(), ctx.ItemTail(then))
}
