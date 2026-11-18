package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_afr_b18_helpers.go — shared seams for Adventures in the Forgotten
// Realms batch 18 (S82).

// afrB18OpponentCreatureDealtDamage is the creature an opponent of the
// delayed trigger's controller controls that a damage event reached, while it
// is a creature on the battlefield (Green Dragon). It is Shriveling Rot's
// each-condition narrowed by the creature's controller as it is dealt the
// damage.
func afrB18OpponentCreatureDealtDamage(ev game.Event, dt *game.DelayedTrigger, g *game.Game, _ game.EffectParams) []uuid.UUID {
	c, ok := b28CreatureWasDealtDamage(ev, g)
	if !ok || dt == nil || c.Controller == dt.Controller {
		return nil
	}
	return []uuid.UUID{c.InstanceID}
}

// An on-disk key: never renamed, never reused.
var afrB18OpponentCreatureDealtDamageCondition = game.DelayedEachCondition("damage/each-opponent-creature-dealt-damage", afrB18OpponentCreatureDealtDamage)

// afrB18OptionalBeholdChosen is OptionalBehold that also records the pick on
// the payment (RevealCost.Chosen), for "you may behold a Dragon … deals damage
// equal to the power of that card or creature instead" (Dragon's Fire): the
// resolution reads the power through b25ChosenPower, a permanent live (or as
// it last existed there, CR 608.2h) and a revealed card as it was revealed.
func afrB18OptionalBeholdChosen(article, subtype string) game.AdditionalCost {
	c := OptionalBehold(article, subtype)
	c.Reveal.Chosen = true
	return c
}

// afrB18TotalPowerOfAttackingCreatures is the sum of the current power of
// every attacking creature on the battlefield, whoever controls it.
func afrB18TotalPowerOfAttackingCreatures(g *game.Game) int {
	total := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.IsCreature() && c.AttackingTarget != uuid.Nil {
			total += c.CurrentPower()
		}
	}
	return total
}

// afrB18PowerOf is the current power of the battlefield permanent `id`, and
// whether it is on the battlefield.
func afrB18PowerOf(g *game.Game, id uuid.UUID) (int, bool) {
	c, ok := battlefieldCardForEffect(g, id)
	if !ok {
		return 0, false
	}
	return c.CurrentPower(), true
}

// afrB18CountersEqualToDamageOnThis is "put that many +1/+1 counters on
// <this creature>" for a trigger whose event is the damage it dealt: the
// counters land only while the source is still the permanent that triggered
// (CR 400.7). Westgate Regent and Abomination, Irradiated Brute share it.
func afrB18CountersEqualToDamageOnThis(ctx *Context) error {
	n := ctx.Trigger().Event.Amount
	if n <= 0 || !sourceIsStillThisPermanent(ctx.Game, ctx.Item) {
		return nil
	}
	return CounterOnThis{Kind: game.CounterPlusOne, N: n}.Apply(ctx)
}

// afrB18DamageEachCreatureWithoutFlying is "deals X damage to each creature
// without flying", X the announced X, as one damage instance. Klauth's Will
// and Subterranean Tremors share it.
func afrB18DamageEachCreatureWithoutFlying(ctx *Context) error {
	x := ctx.X()
	if x <= 0 {
		return nil
	}
	return ctx.Game.DamageInstanceForEffect(func() error {
		return damageEachMatching(ctx, And(Creature(), WithoutKeyword("flying")), x)
	})
}
