package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Swashbuckler's Whip — Artifact — Equipment {1}:
//
//	"Equipped creature has reach, "{2}, {T}: Tap target artifact or
//	 creature," and "{8}, {T}: Discover 10."
//	 Equip {1}"
//
// Reach is a keyword grant; the two quoted abilities are one catalog
// bundle granted to the equipped creature (ADR 0093), so the creature's
// controller activates them and {T} taps the creature. Discover is ADR
// 0099's (game/discover.go).
func init() {
	Register(Spec{
		OracleID:     "5cdb69e5-9395-4ef6-be18-71e3dd70df66",
		Name:         "Swashbuckler's Whip",
		Completeness: CompletenessFull,
		Discovers:    true,
		Grants: []AbilityGrant{{
			Key: swashbucklersWhipGrant,
			Activated: []ActivatedAbility{
				{
					Label:   "{2}, {T}: Tap target artifact or creature",
					Cost:    Plus(ManaCost("{2}"), TapCost()),
					Targets: TargetPermanent("target artifact or creature", Or(Artifact(), Creature())),
					Effect: func(g *game.Game, item *game.StackItem) error {
						ctx := NewContext(g, item)
						for _, t := range ctx.LegalTargets() {
							if t.Kind == game.TargetCard {
								return g.TapTargetForEffect(t.ID)
							}
						}
						return nil
					},
				},
				{
					Label:   "{8}, {T}: Discover 10",
					Purpose: game.Purpose{Answers: game.AnswerValue},
					Cost:    Plus(ManaCost("{8}"), TapCost()),
					Effect:  DiscoverN(10),
				},
			},
			Text: "{2}, {T}: Tap target artifact or creature. {8}, {T}: Discover 10.",
		}},
		Static: []game.StaticAbility{
			GrantToAttached("reach"),
			GrantAbilitiesToAttached(swashbucklersWhipGrant),
		},
		Activated: []ActivatedAbility{
			EquipAbility("{1}"),
		},
	})
}

const swashbucklersWhipGrant = "swashbucklers-whip/tap-and-discover"
