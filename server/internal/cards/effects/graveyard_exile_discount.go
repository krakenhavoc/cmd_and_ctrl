package effects

import (
	"fmt"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// graveyard_exile_discount.go — #5895 (Gorex, the Tombshell): "As an
// additional cost to cast this spell, you may exile any number of creature
// cards from your graveyard. This spell costs {2} less to cast for each card
// exiled this way." and the "exiled with" trigger that reads those cards.
//
// No new wire: the clause is the X form of the cast's graveyard exile
// (ExileXFromGraveyardCost, #1948), whose picker the client already runs.
// The caster announces how many cards (X, CR 601.2b; zero is legal, so the
// cost never makes the spell uncastable) and names exactly that many on
// exile_ids, capped at what the graveyard holds (CR 118.3). "Any number"
// and "announce a number, then name that many" are the same choice. The
// card prints no {X}, so the number adds no mana (CR 107.3) and is not part
// of the spell's mana value (CR 202.3e); its only reader is the discount.

// ExileAnyNumberFromGraveyardLinkedCost is "As an additional cost to cast
// this spell, you may exile any number of <label> from your graveyard",
// with each card linked to the permanent the spell becomes (CR 607.2q):
// the X-form clause with ExileCost.LinkToSource, so the cast's payer stamps
// the spell object and relinkCostExilesLocked moves it to the permanent.
// Pair it with CostsLessPerExiledFromGraveyard, which reads the count, on
// a card whose mana cost prints no {X}. `label` is "X <kind> cards".
func ExileAnyNumberFromGraveyardLinkedCost(label string, match func(game.Card) bool) *game.AdditionalCost {
	ac := ExileXFromGraveyardCost(label, match)
	ac.ExileCards.LinkToSource = true
	return ac
}

// CostsLessPerExiledFromGraveyard is "This spell costs {per} less to cast
// for each card exiled this way" beside ExileAnyNumberFromGraveyardLinkedCost.
// The number exiled IS the announced X (the validator holds the picks to
// exactly X), which CR 601.2b settles before 601.2f totals the cost, so the
// discount is part of the price CastSpell charges, the auto-tap preview
// shows and the enumerator offers, through the one pricer. A reduction
// spends generic mana only and stops at zero (game.reduceGeneric), so Gorex
// always costs at least its {B}{B}. `per` must be generic mana; anything
// else panics at init.
func CostsLessPerExiledFromGraveyard(per string, label string) game.CostModifier {
	unit, err := game.ParseCost(per)
	if err != nil || unit.XSlots > 0 || len(unit.Required) > 0 || unit.Generic < 1 {
		panic(fmt.Sprintf("effects.CostsLessPerExiledFromGraveyard: %q is not a fixed generic amount (%v)", per, err))
	}
	return CostsLessEach(func(q game.CostQuery) int {
		if q.XValue < 1 {
			return 0
		}
		return unit.Generic * q.XValue
	}, label)
}

// ReturnRandomCardExiledWithThis is "choose a card at random exiled with
// <this> and put that card into its owner's hand" (Gorex, the Tombshell).
// The cards are the ones linked to the source OBJECT the trigger came from
// (CR 607.2a, its last-known reference when it has died, CR 608.2h); a card
// that has left exile has lost the link (CR 400.7). None linked: nothing
// happens.
type ReturnRandomCardExiledWithThis struct{}

// Apply draws one linked card through the game's random door and returns it.
func (ReturnRandomCardExiledWithThis) Apply(ctx *Context) error {
	ref, ok := ctx.SourceRef()
	if !ok {
		return nil
	}
	var ids []uuid.UUID
	for _, c := range ctx.Game.ExiledWithForEffect(game.PermissionCardRef(ref)) {
		ids = append(ids, c.InstanceID)
	}
	pick := ctx.Game.ChooseAtRandomForEffect(RandomDraw(ctx), ids, 1)
	if len(pick) == 0 {
		return nil
	}
	return BounceToHand{Target: pick[0]}.Apply(ctx)
}
