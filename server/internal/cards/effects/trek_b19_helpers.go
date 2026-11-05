package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// trek_b19_helpers.go — shared pieces for Star Trek batch 19 (#4973).

// thatCreatureExplores is the body of "whenever a creature you control
// enters, it explores" (Path of Discovery; Captain Kathryn Janeway's
// "Whenever Captain Janeway or another creature you control enters, that
// creature explores"): the object that entered explores under the
// trigger's controller (CR 701.44a). A creature that has left by the time
// the trigger resolves still explores by its last known information (CR
// 701.44c): the card is revealed, and no counter is put on anything.
func thatCreatureExplores(g *game.Game, item *game.StackItem) error {
	if o := item.Trigger.Object; o != nil {
		// The object that entered, not a later one.
		return g.ExploreForEffect(item.SourceCardID, game.ObjectRef{ID: o.ID, Epoch: o.Epoch}, item.Controller)
	}
	return Explores{Explorer: item.Trigger.Event.CardID}.Apply(NewContext(g, item))
}

// trekB19DifferentlyNamedLandsPT is the characteristic-defining ability
// "<this>'s power and toughness are each equal to the number of
// differently named lands you control" (Captain Pike, Fullness of Time;
// CR 604.3a), counted every layer pass: each land its controller
// controls, once per name. It reads the battlefield directly, as
// landsYouControlPT does, so it is safe inside a layer pass.
func trekB19DifferentlyNamedLandsPT(label string) game.StaticAbility {
	return game.StaticAbility{
		Layer:     game.Layer7PT,
		SubLayer:  game.SubLayer7A_CDA,
		Label:     label,
		AppliesTo: selfOnly,
		Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
			n := 0
			if g.Battlefield != nil {
				names := map[string]bool{}
				for i := range g.Battlefield.Cards {
					l := &g.Battlefield.Cards[i]
					if l.Controller == source.Controller && l.IsLand() {
						names[l.Name] = true
					}
				}
				n = len(names)
			}
			c.Power, c.Toughness = n, n
		},
	}
}

// eachDealsDamageEqualToItsPower is "they each deal damage equal to their
// power to <victim>" (Terrific Team-Up, Common Goal): each source still on
// the battlefield deals damage equal to its current power to the victim,
// all in one damage event, each with its own deathtouch and lifelink. The
// caller has already dropped the sources that are no longer legal targets.
func eachDealsDamageEqualToItsPower(ctx *Context, sources []game.TargetRef, victim uuid.UUID) error {
	return ctx.Game.DamageInstanceForEffect(func() error {
		for _, s := range sources {
			c, ok := ctx.Game.LookupCardForEffect(s.ID)
			if !ok {
				continue
			}
			if err := (DealDamage{Source: s.ID, Target: victim, Amount: c.CurrentPower()}).Apply(ctx); err != nil {
				return err
			}
		}
		return nil
	})
}
