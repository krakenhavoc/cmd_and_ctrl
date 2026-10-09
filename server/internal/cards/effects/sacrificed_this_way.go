package effects

import (
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// sacrificed_this_way.go — shared bodies for the spells that read the
// permanent their additional cost sacrificed (ADR 0113 §1, #2072).
// Append-only: a new reader goes at the bottom.
//
// The record is PaidCost.SacrificedObjects, read through
// Context.SacrificedPermanents / SacrificedPower / SacrificedTotalPower:
// each sacrificed permanent as it last existed on the battlefield (CR
// 400.7j, 608.2h), so an anthem's bonus and its counters count, and a
// CR 707.10 copy reads the permanent the ORIGINAL sacrificed.

// damageEqualToSacrificedPower is "deals damage equal to the sacrificed
// creature's power to any target" (Fling, Kazuul's Fury): the first
// target takes the sacrificed creature's last-known power, floored at
// zero (CR 107.1b), so a creature with negative power deals none. A
// target that is gone or illegal by resolution is the engine's CR
// 608.2b re-check, as for every single-target damage spell.
func damageEqualToSacrificedPower(item *game.StackItem, ctx *Context) error {
	if len(item.Targets) == 0 {
		return nil
	}
	return DealDamage{
		Source: ctx.Source(),
		Target: item.Targets[0].ID,
		Amount: ctx.SacrificedPower(),
	}.Apply(ctx)
}

// targetOpponentOrPlaneswalker is "target opponent or planeswalker"
// (Final Strike): an opponent, or a planeswalker on the battlefield.
func targetOpponentOrPlaneswalker() *game.TargetSpec {
	spec := targetPlayerOrPlaneswalker()
	spec.Label = "target opponent or planeswalker"
	spec.PlayerOK = func(g *game.Game, caster uuid.UUID, p *game.Player) bool {
		return Opponent()(g, caster, p)
	}
	return spec
}

// drawEqualToSacrificedPower is "draw cards equal to the sacrificed
// creature's power" (Life's Legacy): its last-known power, and none for
// a power of zero or less (CR 107.1b).
func drawEqualToSacrificedPower(_ *game.StackItem, ctx *Context) error {
	return DrawCards{N: ctx.SacrificedPower()}.Apply(ctx)
}

// sacrificedManaValue is the sacrificed permanent's mana value as it
// last existed on the battlefield (Morbid Curiosity, Reckoner's Bargain,
// Forge Armor): a token that is no copy, or a face-down permanent, is 0.
func sacrificedManaValue(ctx *Context) int {
	info, ok := ctx.SacrificedPermanent()
	if !ok {
		return 0
	}
	return info.ManaValue
}

// sacrificedToughness is the sacrificed creature's toughness as it last
// existed on the battlefield, floored at zero (CR 107.1b) — Severed
// Strands' and Worthy Cause's "life equal to the sacrificed creature's
// toughness".
func sacrificedToughness(ctx *Context) int {
	info, ok := ctx.SacrificedPermanent()
	if !ok || info.Toughness < 0 {
		return 0
	}
	return info.Toughness
}

// sacrificedHadSupertype reports whether the sacrificed permanent had
// the supertype as it last existed on the battlefield — Nasty End's "if
// the sacrificed creature was legendary".
func sacrificedHadSupertype(ctx *Context, supertype string) bool {
	info, ok := ctx.SacrificedPermanent()
	return ok && hasFold(info.Characteristic.Supertypes, supertype)
}

// sacrificedHadCardType reports whether the sacrificed permanent had the
// card type as it last existed on the battlefield — Foundry Helix's "if
// the sacrificed permanent was an artifact".
func sacrificedHadCardType(ctx *Context, cardType string) bool {
	info, ok := ctx.SacrificedPermanent()
	return ok && hasFold(info.Characteristic.Types, cardType)
}

// sacrificedHadSubtype reports whether the sacrificed permanent had the
// subtype as it last existed on the battlefield — Hellish Sideswipe's
// "if the sacrificed permanent was a Vehicle".
func sacrificedHadSubtype(ctx *Context, subtype string) bool {
	info, ok := ctx.SacrificedPermanent()
	return ok && hasFold(info.Characteristic.Subtypes, subtype)
}

// sharesACardTypeWith reports whether `c` shares a card type (CR 205.2a)
// with the sacrificed permanent as it last existed — Fatal Grudge.
// Subtypes and supertypes do not count.
func sharesACardTypeWith(info game.PermanentInfo, c game.Card) bool {
	for _, t := range info.Characteristic.Types {
		if hasFold(cardTypesCR205, t) && c.HasCardType(strings.ToLower(t)) {
			return true
		}
	}
	return false
}

// sharesACreatureTypeWith reports whether `c` shares a creature type
// with the sacrificed creature as it last existed — Endemic Plague. The
// same rule as game.SharesCreatureType, read off a last-known record: a
// changeling on either side shares with anything that has a creature
// type (CR 702.73a), and land and artifact subtypes never count.
func sharesACreatureTypeWith(info game.PermanentInfo, c game.Card) bool {
	theirs := game.CreatureTypesOf(&c)
	if len(theirs) == 0 {
		return false
	}
	var mine []string
	for _, t := range info.Characteristic.Subtypes {
		if game.IsCreatureType(t) {
			mine = append(mine, t)
		}
	}
	if info.Characteristic.AllCreatureTypes || game.HasAllCreatureTypes(&c) {
		return info.Characteristic.AllCreatureTypes || len(mine) > 0
	}
	for _, t := range mine {
		if hasFold(theirs, t) {
			return true
		}
	}
	return false
}

// damageEqualToSacrificedTotalPower is "deals damage to any target
// equal to the total power of the sacrificed creatures" (Soulblast,
// #2097): every sacrificed creature's last-known power summed, the total
// floored at zero (ctx.SacrificedTotalPower, CR 107.1b), to the first
// target. Nothing sacrificed is no damage.
func damageEqualToSacrificedTotalPower(item *game.StackItem, ctx *Context) error {
	if len(item.Targets) == 0 {
		return nil
	}
	return DealDamage{
		Source: ctx.Source(),
		Target: item.Targets[0].ID,
		Amount: ctx.SacrificedTotalPower(),
	}.Apply(ctx)
}
