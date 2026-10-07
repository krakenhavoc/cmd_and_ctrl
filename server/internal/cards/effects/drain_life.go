package effects

import (
	"math"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Drain Life — Sorcery {X}{1}{B}:
//
//	"Spend only black mana on X.
//	 Drain Life deals X damage to any target. You gain life equal to
//	 the damage dealt, but not more life than the player's life total
//	 before the damage was dealt, the planeswalker's loyalty before the
//	 damage was dealt, or the creature's toughness."
//
// #2556: the first card to print a cost clause on the SPELL itself.
// SpellSpendOnlyOnX("B") rides the cast pricer (ADR 0040's 2026-10-07
// amendment), so the cast, its auto-tap, the legal-move list and the
// auto-tap preview all accept black mana for X and nothing else, and
// the price shown stays {X}{1}{B}.
//
// The life gained is what the damage actually landed as (a prevention
// shield or a doubler changes it, and a CR 616 prompt defers it —
// DealDamageThen), capped by the recipient as it stood BEFORE the
// damage: its life total, its loyalty, or its toughness.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e75ba79f-4cc2-4ede-8641-559ab94e7e36",
		Name:         "Drain Life",
		XMatters:     true,
		Completeness: CompletenessFull,
		SpendOnly:    SpellSpendOnlyOnX("B"),
		Targets:      TargetAny(),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return damageThenGainUpToWhatItWas(ctx, item, noExtraLifeCap)
		},
	})
}

// noExtraLifeCap is the extra cap a card with none passes to
// damageThenGainUpToWhatItWas.
const noExtraLifeCap = math.MaxInt

// damageThenGainUpToWhatItWas is the Drain Life family's body: X damage
// to the spell's one target, then its controller gains life equal to the
// damage dealt, but not more than the recipient's life total (a player),
// loyalty (a planeswalker) or toughness (a creature) before the damage,
// and not more than `extraCap` (Soul Burn's black mana spent on X).
// The recipient's number is read BEFORE the damage, because the cap is
// printed that way; the gain waits on the damage's own continuation.
func damageThenGainUpToWhatItWas(ctx *Context, item *game.StackItem, extraCap int) error {
	t, ok := firstLegalTarget(ctx)
	if !ok {
		return nil
	}
	limit := min(extraCap, lifeGainCapBeforeDamage(ctx.Game, t))
	controller, source := item.Controller, ctx.Source()
	return DealDamageThen(ctx, t.ID, ctx.X(), func(g *game.Game, _ uuid.UUID, dealt int) error {
		gain := min(dealt, limit)
		if gain <= 0 {
			return nil
		}
		return g.ChangePlayerLifeForEffect(source, controller, gain)
	})
}

// lifeGainCapBeforeDamage is "the player's life total before the damage
// was dealt, the planeswalker's loyalty before the damage was dealt, or
// the creature's toughness" for one target: whichever applies, and the
// lowest when a permanent is both a planeswalker and a creature.
func lifeGainCapBeforeDamage(g *game.Game, t game.TargetRef) int {
	if p := g.PlayerByIDForEffect(t.ID); p != nil {
		return p.Life
	}
	c, ok := g.LookupCardForEffect(t.ID)
	if !ok {
		return 0
	}
	limit := math.MaxInt
	if c.IsPlaneswalker() {
		limit = min(limit, c.Counters[game.CounterLoyalty])
	}
	if c.IsCreature() {
		limit = min(limit, c.CurrentToughness())
	}
	if limit == math.MaxInt {
		// Neither: the target was not one the clause allows by now
		// (CR 608.2b); the weaker answer is no life.
		return 0
	}
	return limit
}
