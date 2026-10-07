package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// energy_resolution.go — ADR 0129 §3: the shared lines of the cards that
// pay energy as an ability resolves. Append-only: add a builder, never
// change what one means. The primitives are in energy_payment.go.

// whenThisAttacksMayPayEnergy is "Whenever this creature attacks, you
// may pay N {E}. If you do, <then>." — the Thriving cycle, the Aether
// Chaser cycle, Riparian Tiger. `what` finishes the question: "pay {E}{E}
// to <what>?". The prompt holds the declare attackers step, so what is
// bought lands before blockers and damage.
func whenThisAttacksMayPayEnergy(name string, n int, what string, then Effect) game.TriggeredAbility {
	return WheneverThisAttacks(name+" — you may pay "+EnergySymbols(n), mayPayEnergyThen(name, n, what, then))
}

// mayPayEnergyThen is the body of "you may pay N {E}. If you do, <then>"
// as a trigger's effect, for any trigger condition.
func mayPayEnergyThen(name string, n int, what string, then Effect) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		return MayPayEnergy{
			N:        n,
			Question: name + " — pay " + EnergySymbols(n) + " to " + what + "?",
			OnPay: func(ctx *Context) error {
				return then(ctx.Game, ctx.Item)
			},
		}.Apply(NewContext(g, item))
	}
}

// thisStillHere runs `then` only while the source is the same permanent
// that triggered (CR 400.7): "put a +1/+1 counter on it" means the
// creature that attacked, not a new object with its name.
func thisStillHere(then Effect) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		if !sourceIsStillThisPermanent(g, item) {
			return nil
		}
		return then(g, item)
	}
}

// createServo is "create a 1/1 colorless Servo artifact creature token".
func createServo(g *game.Game, item *game.StackItem) error {
	return CreateToken{Template: TokenCard("1/1 colorless Servo artifact"), N: 1}.Apply(NewContext(g, item))
}

// sacrificeThisUnlessYouPayEnergy is "sacrifice this permanent unless
// you pay N {E}" (CR 118.12a) as a trigger's effect. Declining, or being
// short, sacrifices it — the permanent that triggered, not a new object
// (CR 400.7).
func sacrificeThisUnlessYouPayEnergy(name, permanent string, n int) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		return PayEnergyUnless{
			N:        n,
			Question: name + " — pay " + EnergySymbols(n) + ", or sacrifice " + permanent + "?",
			OnDecline: func(ctx *Context) error {
				if !sourceIsStillThisPermanent(ctx.Game, ctx.Item) {
					return nil
				}
				return SacrificePermanent{Target: ctx.Item.SourceCardID}.Apply(ctx)
			},
		}.Apply(NewContext(g, item))
	}
}

// lethalDamageGoal is a PayEnergyAmount.Goal for "deals that much damage
// to that permanent": the damage that destroys the spell's first target —
// a creature's toughness less the damage already marked on it, a
// planeswalker's loyalty (CR 120.3c). Zero names no goal.
func lethalDamageGoal(ctx *Context) int {
	if ctx.Item == nil || len(ctx.Item.Targets) == 0 {
		return 0
	}
	c, ok := ctx.Game.LookupCardForEffect(ctx.Item.Targets[0].ID)
	if !ok {
		return 0
	}
	if !c.IsCreature() {
		return c.Counters[game.CounterLoyalty]
	}
	lethal := c.CurrentToughness() - c.DamageMarked
	if lethal < 0 {
		return 0
	}
	return lethal
}

// getEnergyThenPayAnyAmountToDamageTarget is Harnessed Lightning's and
// Galvanic Discharge's sentence: "Choose target <permanent>. You get N
// {E}, then you may pay any amount of {E}. <This> deals that much damage
// to that <permanent>."
func getEnergyThenPayAnyAmountToDamageTarget(name string, n int) func(item *game.StackItem, ctx *Context) error {
	return func(item *game.StackItem, ctx *Context) error {
		if len(item.Targets) == 0 {
			return nil
		}
		target := item.Targets[0].ID
		if err := (GetEnergy{N: n}).Apply(ctx); err != nil {
			return err
		}
		return PayEnergyAmount{
			Question: name + " — pay any amount of {E} for that much damage",
			Unit:     game.PayAmountDamage,
			Goal:     lethalDamageGoal,
			Then: func(ctx *Context, paid int) error {
				if paid <= 0 {
					return nil
				}
				return DealDamage{Source: ctx.Source(), Target: target, Amount: paid}.Apply(ctx)
			},
		}.Apply(ctx)
	}
}

// voltaicBrawlerPump is Voltaic Brawler's "it gets +1/+1 and gains
// trample until end of turn": two until-end-of-turn effects on the same
// creature, each refusing a new object (CR 400.7).
func voltaicBrawlerPump(g *game.Game, item *game.StackItem) error {
	if err := thisGetsUntilEndOfTurn(1, 1, "Voltaic Brawler — +1/+1 until end of turn")(g, item); err != nil {
		return err
	}
	return thisGainsKeywordUntilEndOfTurn("trample", "Voltaic Brawler — trample until end of turn")(g, item)
}
