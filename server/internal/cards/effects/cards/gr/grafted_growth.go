package cardsgr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Grafted Growth — Enchantment — Aura {2}{G} (Neon Dynasty, #5493):
//
//	"Enchant land
//	 When this Aura enters, put a +1/+1 counter on target creature or
//	 Vehicle you control.
//	 Enchanted land has "{T}: Add two mana of any one color.""
//
// The counter goes on the trigger's target as it resolves (a Vehicle takes
// it whether or not it is crewed). The mana ability is a granted bundle
// (ADR 0093) on the enchanted land: both mana come from the one color the
// land's controller picks, and it ends with the Aura.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c4525b1c-5c3d-494e-b660-ba5a5a28f621",
		Name:         "Grafted Growth",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantLand(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Grants: []effects.AbilityGrant{{
			Key:  graftedGrowthMana,
			Text: "{T}: Add two mana of any one color.",
			Mana: []effects.ManaAbility{{
				Cost:     effects.ManaAbilityCost{Tap: true},
				Produced: effects.OneColorOfAmount(2),
				Label:    "Add two mana of any one color",
			}},
		}},
		Static: []game.StaticAbility{effects.GrantAbilitiesToAttached(graftedGrowthMana)},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Grafted Growth — put a +1/+1 counter on target creature or Vehicle you control",
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						if t := effects.FirstLegalBattlefieldTarget(ctx); t != uuid.Nil {
							return effects.AddCounter{Target: t, Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
						}
						return nil
					}),
				effects.TargetPermanent("target creature or Vehicle you control", effects.Or(effects.Creature(), effects.Subtype("Vehicle")), effects.YouControl())),
		},
	})
}

const graftedGrowthMana = "grafted-growth/mana"
