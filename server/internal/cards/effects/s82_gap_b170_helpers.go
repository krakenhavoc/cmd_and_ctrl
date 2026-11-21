package effects

import (
	"errors"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Helpers for S82 GAP batch 170 (Jumpstart 2022 enchantments and instants,
// #5948).

// GapB170TapTwoSharingATypeCost is Weight of Conscience's cost, "Tap two
// untapped creatures you control that share a creature type".
func GapB170TapTwoSharingATypeCost() game.AbilityCost {
	const label = "two untapped creatures you control that share a creature type"
	return game.AbilityCost{TapOthers: &game.TapOthersCost{
		Count:              2,
		Filter:             TargetCreature(label),
		Label:              label,
		SharesCreatureType: true,
	}}
}

// GapB170ReturnThisToBattlefield is the body of the dies trigger Demonic
// Gifts grants: "return it to the battlefield under its owner's control"
// (untapped, unlike the tapped returns of Fake Your Own Death). A token is
// gone (CR 111.7) and a card that has already left the graveyard is a new
// object (CR 400.7), so neither returns.
func GapB170ReturnThisToBattlefield(g *game.Game, item *game.StackItem) error {
	z := g.FindCardZoneForEffect(item.SourceCardID)
	if z == nil || z.Kind != game.ZoneGraveyard || !SameGraveyardObject(z, item) {
		return nil
	}
	if _, err := g.ReturnFromGraveyardWithCountersForEffect(item.SourceCardID, uuid.Nil, false, nil); err != nil && !errors.Is(err, game.ErrCardNotFound) {
		return err
	}
	return nil
}

// GapB170PlusXCreatureCards is Ghoul's Feast's "+X/+0 until end of turn,
// where X is the number of creature cards in your graveyard", counted as
// the spell resolves.
func GapB170PlusXCreatureCards(item *game.StackItem, ctx *Context) error {
	id, ok := B16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	return BoostUntilEOT{Target: id, Power: CreatureCardsInGraveyard(ctx.Game, item.Controller),
		Label: "Ghoul's Feast — +X/+0 until end of turn"}.Apply(ctx)
}

// GapB170HungerCounters is Hunger of the Howlpack's body: one +1/+1 counter
// on the target, or three if a creature died this turn (morbid), read as it
// resolves.
func GapB170HungerCounters(item *game.StackItem, ctx *Context) error {
	id, ok := B16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	n := 1
	if B11CreaturesDiedThisTurn(ctx.Game) > 0 {
		n = 3
	}
	return AddCounter{Target: id, Kind: game.CounterPlusOne, N: n}.Apply(ctx)
}

// GapB170MinusOneAndSaproling is Fungal Infection's body: the target gets
// -1/-1 until end of turn, and its caster creates a 1/1 green Saproling
// creature token. With its one target illegal the spell does nothing at all,
// token included (CR 608.2b).
func GapB170MinusOneAndSaproling(item *game.StackItem, ctx *Context) error {
	id, ok := B16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	if err := (BoostUntilEOT{Target: id, Power: -1, Toughness: -1, Label: "Fungal Infection — -1/-1 until end of turn"}).Apply(ctx); err != nil {
		return err
	}
	return CreateToken{Template: TokenCard("1/1 green Saproling"), N: 1}.Apply(ctx)
}

// GapB170MinusXMinusX is Death Wind's body: the target gets -X/-X until end
// of turn, X being the announced X (read by the card, which declares
// XMatters).
func GapB170MinusXMinusX(ctx *Context, x int) error {
	id, ok := B16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	if x <= 0 {
		return nil
	}
	return BoostUntilEOT{Target: id, Power: -x, Toughness: -x, Label: "Death Wind — -X/-X until end of turn"}.Apply(ctx)
}
