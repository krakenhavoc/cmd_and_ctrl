package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

const citizensCrowbarGrant = "citizens-crowbar/destroy"

// Citizen's Crowbar — Artifact — Equipment {1}{W} (Streets of New Capenna,
// #5491):
//
//	"When this Equipment enters, create a 1/1 green and white Citizen creature
//	 token, then attach this Equipment to it.
//	 Equipped creature gets +1/+1 and has "{W}, {T}, Sacrifice Citizen's
//	 Crowbar: Destroy target artifact or enchantment."
//	 Equip {2}"
//
// Dwarven Shortsword's token-and-attach entry trigger, and Deconstruction
// Hammer's grant: the quoted ability is the equipped creature's own (so it
// taps to pay, CR 302.6), and "Sacrifice Citizen's Crowbar" is the
// grantor-sacrifice cost.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "64c7e03d-ba38-4979-8889-9bf0e7b8c66e",
		Name:         "Citizen's Crowbar",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key: citizensCrowbarGrant,
			Activated: []effects.ActivatedAbility{{
				Label:   "{W}, {T}, Sacrifice Citizen's Crowbar: Destroy target artifact or enchantment.",
				Cost:    effects.Plus(effects.ManaCost("{W}"), effects.TapCost(), effects.SacrificeTheGrantor()),
				Targets: effects.TargetPermanent("target artifact or enchantment", effects.Or(effects.Artifact(), effects.Enchantment())),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.DestroyTheTargetPermanent(item, effects.NewContext(g, item))
				},
			}},
			Text: "{W}, {T}, Sacrifice Citizen's Crowbar: Destroy target artifact or enchantment.",
		}},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Citizen's Crowbar — create a 1/1 green and white Citizen creature token, then attach this Equipment to it",
				func(g *game.Game, item *game.StackItem) error {
					return effects.CreateTokenAndAttachSource(g, item, "1/1 green and white Citizen", true)
				}),
		},
		Static: []game.StaticAbility{
			effects.PumpAttached(1, 1),
			effects.GrantAbilitiesToAttached(citizensCrowbarGrant),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{2}")},
	})
}
