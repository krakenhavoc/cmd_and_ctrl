package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// life_gated_statics.go — the shared shape of "as long as <somebody>
// has N or more / fewer life, …" (#1117).
//
// A static whose condition reads a LIFE TOTAL has one requirement no
// other static has: it must declare StaticAbility.DependsOnLifeTotal,
// or the layer engine's cached resolution survives the life change
// and the card answers one event late. That is not something a card
// author should have to remember, so every constructor here sets the
// flag itself and the cards call these rather than building the
// StaticAbility literal. See invalidateLayersForLifeChangeLocked in
// internal/game for what the flag buys and why the bump it gates
// rides the WRITE to a life total rather than an event.
//
// Append new shapes here; never change one that exists.

// LifeGatedPump is the layer-7c half: "<these permanents> get +p/+t"
// while some life-total condition holds. Serra Ascendant's +5/+5,
// Righteous Valkyrie's +2/+2.
func LifeGatedPump(applies func(target *game.Card, g *game.Game, source *game.Card) bool, power, toughness int) game.StaticAbility {
	return game.StaticAbility{
		Layer:              game.Layer7PT,
		SubLayer:           game.SubLayer7C_Modify,
		DependsOnLifeTotal: true,
		AppliesTo:          applies,
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
			c.Power += power
			c.Toughness += toughness
		},
	}
}

// LifeGatedKeyword is the layer-6 half: "<these permanents> have
// <keyword>" while some life-total condition holds. Serra Ascendant's
// flying, Bloodghast's haste.
//
// Built on KeywordGrant so the append-if-absent dedupe is the same
// one every other grant uses.
func LifeGatedKeyword(applies func(target *game.Card, g *game.Game, source *game.Card) bool, keyword string) game.StaticAbility {
	ab := KeywordGrant(applies, keyword)
	ab.DependsOnLifeTotal = true
	return ab
}

// SelfWhileYourLifeAtLeast — "as long as you have N or more life,
// THIS creature …" (Serra Ascendant).
func SelfWhileYourLifeAtLeast(n int) func(target *game.Card, g *game.Game, source *game.Card) bool {
	return func(target *game.Card, g *game.Game, source *game.Card) bool {
		return target.InstanceID == source.InstanceID && controllerLifeAtLeast(g, source, n)
	}
}

// YourCreaturesWhileYourLifeAboveStarting — "as long as you have at
// least N life more than your starting life total, creatures you
// control …" (Righteous Valkyrie).
//
// The starting total is the TABLE's (ADR 0075), not the constant: a
// game configured for 20 puts the line at 27, which is what the card
// says.
func YourCreaturesWhileYourLifeAboveStarting(n int) func(target *game.Card, g *game.Game, source *game.Card) bool {
	return func(target *game.Card, g *game.Game, source *game.Card) bool {
		if !target.IsCreature() || target.Controller != source.Controller {
			return false
		}
		return controllerLifeAtLeast(g, source, startingLifeOf(g)+n)
	}
}

// SelfWhileAnOpponentsLifeAtMost — "as long as an opponent has N or
// less life, THIS creature …" (Bloodghast). Any one opponent is
// enough, and an eliminated seat is not an opponent.
func SelfWhileAnOpponentsLifeAtMost(n int) func(target *game.Card, g *game.Game, source *game.Card) bool {
	return func(target *game.Card, g *game.Game, source *game.Card) bool {
		if target.InstanceID != source.InstanceID || g == nil {
			return false
		}
		for _, p := range g.Seats {
			if p == nil || p.ID == source.Controller || p.Eliminated {
				continue
			}
			if p.Life <= n {
				return true
			}
		}
		return false
	}
}

// controllerLifeAtLeast is the one life read the constructors share.
func controllerLifeAtLeast(g *game.Game, source *game.Card, n int) bool {
	if g == nil || source == nil {
		return false
	}
	p := g.PlayerByIDForEffect(source.Controller)
	return p != nil && p.Life >= n
}

// startingLifeOf is the table's configured starting total, falling
// back to the format default for a fixture that never went through
// Start.
func startingLifeOf(g *game.Game) int {
	if g == nil || g.Settings.StartingLife <= 0 {
		return game.StartingLife
	}
	return g.Settings.StartingLife
}
