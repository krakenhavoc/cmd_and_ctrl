package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Golden-Tail Trainer — Creature — Fox Samurai {1}{G}{W}, 1/3 (Modern Horizons 3):
//
//	"Aura and Equipment spells you cast cost {X} less to cast, where X is this
//	 creature's power.
//	 Whenever this creature attacks, other modified creatures you control get
//	 +X/+X until end of turn, where X is this creature's power. (Equipment,
//	 Auras you control, and counters are modifications.)"
//
// X is the Trainer's power as the cost is worked out, and as the attack ability
// resolves (last-known if it has left). The discount lowers generic mana only.
func init() {
	const label = "Golden-Tail Trainer — other modified creatures you control get +X/+X until end of turn, where X is its power"
	effects.Register(effects.Spec{
		OracleID:     "9e0461e7-ccc4-413b-be80-83311be46861",
		Name:         "Golden-Tail Trainer",
		Completeness: effects.CompletenessFull,
		CostModifiers: []game.CostModifier{
			effects.CostsLessEach(func(q game.CostQuery) int { return max(q.Source.PowerForComparison(), 0) },
				"Aura and Equipment spells you cast cost {X} less to cast, where X is this creature's power.",
				effects.YourSpell(), effects.AuraOrEquipmentSpell()),
		},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks(label, func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				info, ok := ctx.SourcePermanent()
				if !ok || info.Power <= 0 {
					return nil
				}
				for _, id := range effects.PermanentsControlledByMatching(g, item.Controller, effects.And(effects.Creature(), effects.ModifiedCreature)) {
					if id == item.SourceCardID {
						continue
					}
					if err := (effects.BoostUntilEOT{Target: id, Power: info.Power, Toughness: info.Power, Label: label}).Apply(ctx.AsGroupMember()); err != nil {
						return err
					}
				}
				return nil
			}),
		},
	})
}
