package cardskr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Krovikan Plague — Enchantment — Aura {2}{B} (Alliances, #5634):
//
//	"Enchant non-Wall creature you control
//	 When this Aura enters, draw a card at the beginning of the next turn's
//	 upkeep.
//	 Tap enchanted creature: This Aura deals 1 damage to any target. Put a -0/-1
//	 counter on enchanted creature. Activate only if enchanted creature is
//	 untapped."
//
// The enters trigger schedules the delayed draw (CR 603.7a). The ability's cost
// is "Tap enchanted creature" (a tap-others cost on the host, not {T}); the Aura
// itself deals the damage, then a -0/-1 counter goes on the enchanted creature
// as it is then (a creature the Aura left has none to take it).
//
// No simplification.
func init() {
	const name = "Krovikan Plague"
	effects.Register(effects.Spec{
		OracleID:     "609c3b1c-d9ad-4616-b685-52b85528fde8",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(effects.YouControl(), effects.NonWall()),
		Purpose:      game.Purpose{Aura: game.AuraBenefit, Draws: 1},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters(name+" — draw a card at the beginning of the next turn's upkeep", effects.Do(effects.DrawACardAtTheNextUpkeep(name+" — draw a card"))),
		},
		Activated: []effects.ActivatedAbility{{
			Label: "Tap enchanted creature: This Aura deals 1 damage to any target. Put a -0/-1 counter on enchanted creature. Activate only if enchanted creature is untapped.",
			Cost:  effects.OldB816TapEnchantedCreature(),
			Condition: func(g *game.Game, _, source uuid.UUID) bool {
				return effects.OldB816EnchantedCreatureIsUntapped(g, source)
			},
			Targets: effects.TargetAny(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if err := effects.DealDamageToTheTarget(1)(g, item); err != nil {
					return err
				}
				ctx := effects.NewContext(g, item)
				host, ok := effects.EnchantedCreatureAsItResolves(ctx)
				if !ok {
					return nil
				}
				return effects.AddCounter{Target: host, Kind: "-0/-1", N: 1}.Apply(ctx)
			},
		}},
	})
}
