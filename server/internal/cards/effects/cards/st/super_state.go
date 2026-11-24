package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Super State — Legendary Enchantment — Aura {7}:
//
//	"Enchant creature you control
//	 Enchanted creature has base power and toughness 9/9 and has flying, first
//	 strike, trample, and haste.
//	 Whenever enchanted creature deals combat damage to an opponent, it deals
//	 that much damage to each other opponent."
//
// Almost Perfect's base-P/T Aura (SetAttachedBasePT, layer 7b, so counters and
// pumps land on top of the 9/9) with four keywords granted to the host at layer
// 6. The trigger is the Aura's own, matched through the attachment, and is
// Hydra Omnivore's: "that much" is the damage the event says was dealt, not the
// creature's power, and the other opponents take it as one damage instance from
// the enchanted creature (so its lifelink or deathtouch applies, and it is not
// combat damage, so it can't trigger again). The creature is the source even
// if it has left by resolution. "An opponent" is an opponent of the Aura's
// controller.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ab2dfe99-bd77-44d7-92be-81643ea901f0",
		Name:         "Super State",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(effects.YouControl()),
		Static: []game.StaticAbility{
			effects.SetAttachedBasePT(9, 9),
			effects.GrantToAttached("flying", "first strike", "trample", "haste"),
		},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventDealDamage},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.Kind != game.EventDealDamage || !ev.Combat || ev.Amount <= 0 || !source.IsAttachedTo(ev.Source) {
					return false
				}
				p := g.PlayerByIDForEffect(ev.Target)
				return p != nil && p.ID != source.Controller
			},
			Key:    "Super State — the enchanted creature deals that much damage to each other opponent",
			Effect: superStateSplash,
		}},
	})
}

func superStateSplash(g *game.Game, item *game.StackItem) error {
	ev := item.Trigger.Event
	ctx := effects.NewContext(g, item)
	return g.DamageInstanceForEffect(func() error {
		for _, opp := range ctx.Opponents() {
			if opp == ev.Target {
				continue
			}
			if err := (effects.DealDamage{Source: ev.Source, Target: opp, Amount: ev.Amount}).Apply(ctx); err != nil {
				return err
			}
		}
		return nil
	})
}
