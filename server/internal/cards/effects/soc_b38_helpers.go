package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// soc_b38_helpers.go — shared pieces for the thirty-eighth Secrets of
// Strixhaven batch (#4456).

// socB38InfusionAppliesToCreaturesYouControl is the AppliesTo of an Infusion
// static (an ability word, CR 207.2c) that reads "creatures you control ...
// as long as you gained life this turn" (Thornfist Striker): a creature its
// controller controls while the controller has gained life this turn. The
// life gained is the turn tally, so the condition is ON as the first point of
// life lands and OFF as the turn changes.
func socB38InfusionAppliesToCreaturesYouControl(target *game.Card, g *game.Game, source *game.Card) bool {
	return target.IsCreature() && target.Controller == source.Controller &&
		b15LifeGainedThisTurn(g, source.Controller) > 0
}

// socB38InfusionAppliesToThis is the AppliesTo of an Infusion static on one
// creature ("this creature gets +2/+0 as long as you gained life this
// turn", Ulna Alley Shopkeep).
func socB38InfusionAppliesToThis(target *game.Card, g *game.Game, source *game.Card) bool {
	return target.InstanceID == source.InstanceID &&
		b15LifeGainedThisTurn(g, source.Controller) > 0
}

// socB38InfusionStatic is b16Anthem or b16GrantKeywords made live for the
// life-gain condition: DependsOnLifeTotal is the layer cache's hint that the
// static has to be recomputed as a life total moves (the tally changes with
// it); the turn change already drops the cache for every table.
func socB38InfusionStatic(st game.StaticAbility, label string) game.StaticAbility {
	st.Label = label
	st.DependsOnLifeTotal = true
	return st
}

// socB38DamageToFirstTarget is damageToFirstTarget as a step body for
// ResolveSteps.
func socB38DamageToFirstTarget(amount int) func(ctx *Context) error {
	return func(ctx *Context) error {
		if len(ctx.Item.Targets) == 0 {
			return nil
		}
		return DealDamage{
			Source: ctx.Source(),
			Target: ctx.Item.Targets[0].ID,
			Amount: amount,
		}.Apply(ctx)
	}
}

// socB38BindToLife is Bind to Life's whole text: "Mill seven cards. Then put
// a creature card from among them onto the battlefield." The pick waits on
// the mill (it can pause on a replacement prompt) as its keyed tail.
func socB38BindToLife(_ *game.StackItem, ctx *Context) error {
	return MillToZone{N: 7, ThenTail: ctx.ItemTail(socB38BindToLifeMilled)}.Apply(ctx)
}

// socB38BindToLifeMilled is the clause after the mill. "From among them" is
// the cards the mill actually put into the graveyard that are still there and
// are creature cards; the controller must pick one when there are any (a lone
// candidate comes back outright).
var socB38BindToLifeMilled = CardTail("bind-to-life/after-mill", func(ctx *Context, o game.TailOutcome) error {
	var creatures []uuid.UUID
	for _, id := range o.IDs {
		if z := ctx.Game.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneGraveyard {
			continue
		}
		if c, ok := ctx.Game.LookupCardForEffect(id); ok && c.IsCreature() {
			creatures = append(creatures, id)
		}
	}
	return b17ReanimateOneMilled(ctx.Game, ctx.Item, ctx.Controller(),
		"Bind to Life — put a creature card milled this way onto the battlefield", creatures)
})
