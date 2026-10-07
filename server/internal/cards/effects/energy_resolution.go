package effects

import (
	"github.com/google/uuid"

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

// anotherAttackerMayPayEnergy is "Whenever this creature attacks, you
// may pay {E}. If you do, another target attacking creature <then>"
// (Consul's Shieldguard, Eddytrail Hawk, Smelted Chargebug). The target
// is chosen as the trigger goes on the stack (CR 603.3d); the payment is
// made as it resolves.
func anotherAttackerMayPayEnergy(name, what string, then Effect) game.TriggeredAbility {
	t := whenThisAttacksMayPayEnergy(name, 1, what, then)
	t.Targets = Another(TargetCreature("another target attacking creature", AttackingCreature()))
	return t
}

// firstTargetGainsUntilEndOfTurn is "<the target> gets +p/+0 and gains
// <keywords> until end of turn", pinned to the trigger's first target.
// Either half may be empty.
func firstTargetGainsUntilEndOfTurn(power int, label string, keywords ...string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
			return nil
		}
		ctx := NewContext(g, item)
		target := item.Targets[0].ID
		if err := (BoostUntilEOT{Target: target, Power: power, Label: label}).Apply(ctx); err != nil {
			return err
		}
		return GrantKeywordUntilEOT{Target: target, Keywords: keywords, Label: label}.Apply(ctx)
	}
}

// firstTargetManaValue is the mana value of the resolving item's first
// target, wherever it is, for "an amount of {E} equal to that
// permanent's mana value". A target gone or unreadable is 0.
func firstTargetManaValue(g *game.Game, item *game.StackItem) int {
	if item == nil || len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
		return 0
	}
	c, ok := g.LookupCardForEffect(item.Targets[0].ID)
	if !ok {
		return 0
	}
	mv, ok := g.ManaValueForEffect(c)
	if !ok {
		return 0
	}
	return mv
}

// powerThatSavesMostOfYours is Localized Destruction's Goal: the power,
// 1 or more and within the controller's energy, shared by the most
// creatures they control (the greater power on a tie). Zero when none.
func powerThatSavesMostOfYours(ctx *Context) int {
	energy := game.PlayerEnergy(ctx.Game.PlayerByIDForEffect(ctx.Controller()))
	counts := map[int]int{}
	for _, c := range ctx.Game.Battlefield.Cards {
		if c.Controller == ctx.Controller() && c.IsCreature() {
			if p := c.CurrentPower(); p >= 1 && p <= energy {
				counts[p]++
			}
		}
	}
	best, bestN := 0, 0
	for p, n := range counts {
		if n > bestN || (n == bestN && p > best) {
			best, bestN = p, n
		}
	}
	return best
}

// largestOpposingManaValue is Wrath of the Skies' Goal: the largest mana
// value among the artifacts, creatures and enchantments the controller's
// opponents control, capped by the controller's energy. Zero when none.
func largestOpposingManaValue(ctx *Context) int {
	energy := game.PlayerEnergy(ctx.Game.PlayerByIDForEffect(ctx.Controller()))
	best := 0
	for _, c := range ctx.Game.Battlefield.Cards {
		if c.Controller == ctx.Controller() || !(c.IsArtifact() || c.IsCreature() || c.IsEnchantment()) {
			continue
		}
		if mv, ok := ctx.Game.ManaValueForEffect(c); ok && mv > best && mv <= energy {
			best = mv
		}
	}
	return best
}

// putChosenCounterOn is "put your choice of a <kind>, <kind> or <kind>
// counter on <permanent>": an option pick, then one counter of the
// chosen kind, placed by `by`, on `target` if it is still on the
// battlefield.
func putChosenCounterOn(ctx *Context, by, target uuid.UUID, question string, kinds []string) error {
	options := make([]game.ChoiceOption, len(kinds))
	for i, kind := range kinds {
		options[i] = game.ChoiceOption{Label: "A " + kind + " counter"}
	}
	return PickOption{
		Question: question,
		Options:  options,
		Then: func(ctx *Context, index int) error {
			if index < 0 || index >= len(kinds) || !onBattlefield(ctx.Game, target) {
				return nil
			}
			return ctx.Game.AddCounterByForEffect(by, target, kinds[index], 1)
		},
	}.Apply(ctx)
}

// destroyPayloadPermanents destroys each permanent a reflexive trigger
// carries in its payload — "when you do, destroy that permanent" — that
// is still on the battlefield.
func destroyPayloadPermanents(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, id := range ctx.PayloadCards() {
		if !onBattlefield(g, id) {
			continue
		}
		if err := (DestroyTarget{Target: id}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}

// upToOne turns a per-trigger target clause into its "up to one" form:
// choosing no target is legal.
func upToOne(clause func(game.TriggerContext, *game.Card, *game.Game) *game.TargetSpec) func(game.TriggerContext, *game.Card, *game.Game) *game.TargetSpec {
	return func(tc game.TriggerContext, source *game.Card, g *game.Game) *game.TargetSpec {
		spec := clause(tc, source, g)
		if spec != nil {
			spec.Min = 0
			spec.Label = "up to one " + spec.Label
		}
		return spec
	}
}
