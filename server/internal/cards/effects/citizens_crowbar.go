package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "64c7e03d-ba38-4979-8889-9bf0e7b8c66e",
		Name:         "Citizen's Crowbar",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key: citizensCrowbarGrant,
			Activated: []ActivatedAbility{{
				Label:   "{W}, {T}, Sacrifice Citizen's Crowbar: Destroy target artifact or enchantment.",
				Cost:    Plus(ManaCost("{W}"), TapCost(), SacrificeTheGrantor()),
				Targets: TargetPermanent("target artifact or enchantment", Or(Artifact(), Enchantment())),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return destroyTheTargetPermanent(item, NewContext(g, item))
				},
			}},
			Text: "{W}, {T}, Sacrifice Citizen's Crowbar: Destroy target artifact or enchantment.",
		}},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Citizen's Crowbar — create a 1/1 green and white Citizen creature token, then attach this Equipment to it",
				func(g *game.Game, item *game.StackItem) error {
					return createTokenAndAttachSource(g, item, "1/1 green and white Citizen", true)
				}),
		},
		Static: []game.StaticAbility{
			PumpAttached(1, 1),
			GrantAbilitiesToAttached(citizensCrowbarGrant),
		},
		Activated: []ActivatedAbility{EquipAbility("{2}")},
	})
}
