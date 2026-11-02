package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Biorganic Carapace — Artifact — Equipment {2}{W}{U} (Marvel's
// Spider-Man):
//
//	"When this Equipment enters, attach it to target creature you control.
//	 Equipped creature gets +2/+2 and has "Whenever this creature deals
//	 combat damage to a player, draw a card for each modified creature you
//	 control." (Equipment, Auras you control, and counters are
//	 modifications.)
//	 Equip {2}"
//
// Vibranium Strike Gauntlets' enters-attached trigger and granted combat
// trigger (ADR 0093). The draw is the equipped creature's own, so its
// controller draws; the count is made as the trigger resolves (CR 608.2h),
// over the creatures that controller controls that are modified (CR 700.9,
// b07IsModified: a counter, an Equipment, or an Aura they control).
//
// No simplification.
const biorganicCarapaceGrant = "biorganic-carapace/combat-draw"

func init() {
	Register(Spec{
		OracleID:     "7d27c583-469c-42d7-a547-8c65654e9d95",
		Name:         "Biorganic Carapace",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key: biorganicCarapaceGrant,
			Triggered: []game.TriggeredAbility{
				WheneverThisDealsCombatDamageToAPlayer("Biorganic Carapace — draw a card for each modified creature you control",
					func(g *game.Game, item *game.StackItem) error {
						ctx := NewContext(g, item)
						n := countControlled(g, item.Controller, func(c game.Card) bool {
							return c.IsCreature() && b07IsModified(g, c)
						})
						if n == 0 {
							return nil
						}
						return DrawCards{N: n}.Apply(ctx)
					}),
			},
			Text: "Whenever this creature deals combat damage to a player, draw a card for each modified creature you control.",
		}},
		Static: []game.StaticAbility{
			PumpAttached(2, 2),
			GrantAbilitiesToAttached(biorganicCarapaceGrant),
		},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Biorganic Carapace — attach it to target creature you control", AttachSourceToTarget),
				PermanentYouControl("target creature you control", Creature()),
			),
		},
		Activated: []ActivatedAbility{EquipAbility("{2}")},
	})
}
