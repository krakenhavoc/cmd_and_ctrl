package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Syr Armont, the Redeemer — Legendary Creature — Human Knight
// {3}{G}{W}, 4/4:
//
//	"When Syr Armont enters, create a Monster Role token attached to
//	 another target creature you control. (If you control another Role
//	 on it, put that one into the graveyard. Enchanted creature gets
//	 +1/+1 and has trample.)
//	 Enchanted creatures you control get +1/+1."
//
// The anthem is a layer 7c modify on every creature the controller
// controls that has an Aura attached, Syr Armont included if something
// enchants him. It is re-read each layer pass, so it starts and stops
// with the Aura rather than at attach time. The Monster Role's own
// +1/+1 stacks with it on its host.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "53fc09e1-1e5d-4f22-b623-2c5c770183ab",
		Name:         "Syr Armont, the Redeemer",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Syr Armont, the Redeemer — create a Monster Role token attached to another target creature you control",
					createRoleOnFirstTarget(RoleMonster)),
				Another(TargetCreature("another target creature you control", YouControl()))),
		},
		Static: []game.StaticAbility{{
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7C_Modify,
			AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
				return target.IsCreature() && target.Controller == source.Controller && isEnchantedByAura(g, target.InstanceID)
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Power++
				c.Toughness++
			},
		}},
	})
}
