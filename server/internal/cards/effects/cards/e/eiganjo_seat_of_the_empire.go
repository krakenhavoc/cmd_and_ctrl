package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Eiganjo, Seat of the Empire — Legendary Land:
//
//	"{T}: Add {W}.
//	 Channel — {2}{W}, Discard this card: It deals 4 damage to target
//	 attacking or blocking creature. This ability costs {1} less to
//	 activate for each legendary creature you control."
//
// The channel shape Boseiju, Who Endures built (#660): an activated
// ability whose `Zones` is the hand, with `DiscardSelf` as a cost
// component and the legendary-creature discount as the ability's own
// cost clause. The target clause is `AttackingOrBlocking`, which reads
// the combat designations, so it is legal only during a combat that has
// attackers or blockers. "It" is the land: the damage source is the
// discarded card, read off the stack item.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7edb3d15-4f70-4ebe-8c5e-caf6a225076d",
		Name:         "Eiganjo, Seat of the Empire",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{W}",
			Label:    "Add {W}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:         "Channel — {2}{W}, Discard this card: It deals 4 damage to target attacking or blocking creature",
			Cost:          game.AbilityCost{Mana: "{2}{W}", DiscardSelf: true},
			Zones:         []game.ZoneKind{game.ZoneHand},
			CostModifiers: []game.CostModifier{effects.ChannelDiscountPerLegendaryCreature()},
			Targets:       effects.TargetCreature("target attacking or blocking creature", effects.AttackingOrBlocking()),
			Purpose:       effects.ForTargets(effects.DamageToTarget(0, 4)),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				legal := ctx.LegalTargets()
				if len(legal) == 0 {
					return nil
				}
				return effects.DealDamage{Source: item.SourceCardID, Target: legal[0].ID, Amount: 4}.Apply(ctx)
			},
		}},
	})
}
