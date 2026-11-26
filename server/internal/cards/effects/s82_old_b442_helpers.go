package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dark Ascension batch 442 helpers.

// OldB442PumpAttachedWhile is "As long as equipped creature is <X>, it gets
// an additional +p/+t": a layer 7c modify over the attached permanent that
// holds only while `cond` does. `cond` is re-read on every layer pass, so a
// creature that becomes (or stops being) the type gains (loses) the bonus
// without the Equipment moving (Heavy Mattock).
func OldB442PumpAttachedWhile(cond func(host *game.Card, g *game.Game, source *game.Card) bool, power, toughness int) game.StaticAbility {
	return game.StaticAbility{
		Label:    "Attached creature gets " + SignedPT(power, toughness) + " while a condition holds.",
		Layer:    game.Layer7PT,
		SubLayer: game.SubLayer7C_Modify,
		AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
			return AttachedToSource(target, g, source) && cond(target, g, source)
		},
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
			c.Power += power
			c.Toughness += toughness
		},
	}
}

// OldB442AlphaBrawl is Alpha Brawl's resolution: the target creature deals
// damage equal to its power to each other creature its controller controls,
// then each of those creatures deals damage equal to its power to it. Each
// half is one simultaneous damage event (CR 120.2). The creatures are the ones
// found as the spell resolves (CR 608.2b), so a creature the first half kills
// still deals its damage back: state-based actions wait until the spell is done.
func OldB442AlphaBrawl(_ *game.StackItem, ctx *Context) error {
	id, ok := B16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	target, found := ctx.Game.LookupCardForEffect(id)
	if !found || !target.IsCreature() {
		return nil
	}
	owner := target.Controller
	type brawler struct {
		id    uuid.UUID
		power int
	}
	var others []brawler
	for _, c := range MatchingBattlefield(ctx, And(Creature(), NotSelf(id))) {
		if c.Controller == owner {
			others = append(others, brawler{c.InstanceID, c.CurrentPower()})
		}
	}
	if err := ctx.Game.DamageInstanceForEffect(func() error {
		power := target.CurrentPower()
		if power <= 0 {
			return nil
		}
		for _, o := range others {
			if err := (DealDamage{Source: id, Target: o.id, Amount: power}).Apply(ctx); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return err
	}
	return ctx.Game.DamageInstanceForEffect(func() error {
		for _, o := range others {
			if o.power <= 0 {
				continue
			}
			if err := (DealDamage{Source: o.id, Target: id, Amount: o.power}).Apply(ctx); err != nil {
				return err
			}
		}
		return nil
	})
}
