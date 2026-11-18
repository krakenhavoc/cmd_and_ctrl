package effects

import (
	"github.com/google/uuid"

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
	Register(Spec{
		OracleID:     "c4525b1c-5c3d-494e-b660-ba5a5a28f621",
		Name:         "Grafted Growth",
		Completeness: CompletenessFull,
		Targets:      EnchantLand(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Grants: []AbilityGrant{{
			Key:  graftedGrowthMana,
			Text: "{T}: Add two mana of any one color.",
			Mana: []ManaAbility{{
				Cost:     ManaAbilityCost{Tap: true},
				Produced: OneColorOfAmount(2),
				Label:    "Add two mana of any one color",
			}},
		}},
		Static: []game.StaticAbility{GrantAbilitiesToAttached(graftedGrowthMana)},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Grafted Growth — put a +1/+1 counter on target creature or Vehicle you control",
					func(g *game.Game, item *game.StackItem) error {
						ctx := NewContext(g, item)
						if t := FirstLegalBattlefieldTarget(ctx); t != uuid.Nil {
							return AddCounter{Target: t, Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
						}
						return nil
					}),
				TargetPermanent("target creature or Vehicle you control", Or(Creature(), Subtype("Vehicle")), YouControl())),
		},
	})
}

const graftedGrowthMana = "grafted-growth/mana"
