package cardsa

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ajani Unyielding — Legendary Planeswalker — Ajani {4}{G}{W}, starting
// loyalty 4:
//
//	"+2: Reveal the top three cards of your library. Put all nonland
//	 permanent cards revealed this way into your hand and the rest on
//	 the bottom of your library in any order.
//	 −2: Exile target creature. Its controller gains life equal to its
//	 power.
//	 −9: Put five +1/+1 counters on each creature you control and five
//	 loyalty counters on each other planeswalker you control."
//
// The +2 reveals three and takes every nonland permanent card among them
// (no choice), the rest going to the bottom in the order its controller
// picks. The −2 reads the target's controller and power before the
// exile (CR 608.2h) and gives that player the life afterwards, as
// Swords to Plowshares does. The −9 reads the creatures you control and
// the other planeswalkers you control once, as it resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "8a70ddc5-bb51-45eb-96f2-8ecaecac0cc2",
		Name:            "Ajani Unyielding",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 4,
		Activated: []effects.ActivatedAbility{
			{
				Label: "+2: Reveal the top three cards of your library. Put all nonland permanent cards revealed this way into your hand and the rest on the bottom of your library in any order.",
				Cost:  effects.LoyaltyCost(2),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					player := ctx.Controller()
					return effects.TakeFromLibraryToHand{
						Player: player,
						Cards:  g.LookAtTopOfLibraryForEffect(player, 3),
						Match:  effects.And(effects.Nonland(), effects.Permanent()),
						All:    true,
						Label:  "Ajani Unyielding — reveal the top three cards, nonland permanents to your hand and the rest on the bottom",
						Key:    effects.TakeRevealRestBottomAny,
					}.Apply(ctx)
				},
			},
			{
				Label:   "−2: Exile target creature. Its controller gains life equal to its power.",
				Cost:    effects.LoyaltyCost(-2),
				Targets: effects.TargetCreature("target creature"),
				Purpose: game.Purpose{Targets: game.ForTargets(effects.RemovalReturning(0, game.TargetReturn{LifeEqualToPower: true}))},
				Effect:  swordsToPlowsharesOnResolveAbility,
			},
			{
				Label:  "−9: Put five +1/+1 counters on each creature you control and five loyalty counters on each other planeswalker you control.",
				Cost:   effects.LoyaltyCost(-9),
				Effect: ajaniUnyieldingCounters,
			},
		},
	})
}

// swordsToPlowsharesOnResolveAbility adapts Swords to Plowshares' body
// (target exiled, its controller gains life equal to its power) to an
// ability's (g, item) shape.
func swordsToPlowsharesOnResolveAbility(g *game.Game, item *game.StackItem) error {
	return effects.SwordsToPlowsharesOnResolve(item, effects.NewContext(g, item))
}

// ajaniUnyieldingCounters is the −9: five +1/+1 counters on each
// creature the activator controls and five loyalty counters on each
// planeswalker they control other than this Ajani, the set read once as
// it resolves.
func ajaniUnyieldingCounters(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item).AsGroupMember()
	var creatures, walkers []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller != item.Controller {
			continue
		}
		if c.IsCreature() {
			creatures = append(creatures, c.InstanceID)
		}
		if c.IsPlaneswalker() && c.InstanceID != item.SourceCardID {
			walkers = append(walkers, c.InstanceID)
		}
	}
	for _, id := range creatures {
		if err := (effects.AddCounter{Target: id, Kind: game.CounterPlusOne, N: 5}).Apply(ctx); err != nil {
			return err
		}
	}
	for _, id := range walkers {
		if err := (effects.AddCounter{Target: id, Kind: game.CounterLoyalty, N: 5}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}
