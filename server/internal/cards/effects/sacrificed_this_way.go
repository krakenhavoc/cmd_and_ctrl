package effects

import (
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
