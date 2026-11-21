package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "5cdb69e5-9395-4ef6-be18-71e3dd70df66",
		Name:         "Swashbuckler's Whip",
		Completeness: effects.CompletenessFull,
		Discovers:    true,
		Grants: []effects.AbilityGrant{{
			Key: swashbucklersWhipGrant,
			Activated: []effects.ActivatedAbility{
				{
					Label:   "{2}, {T}: Tap target artifact or creature",
					Cost:    effects.Plus(effects.ManaCost("{2}"), effects.TapCost()),
					Targets: effects.TargetPermanent("target artifact or creature", effects.Or(effects.Artifact(), effects.Creature())),
					Effect: func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
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
					Cost:    effects.Plus(effects.ManaCost("{8}"), effects.TapCost()),
					Effect:  effects.DiscoverN(10),
				},
			},
			Text: "{2}, {T}: Tap target artifact or creature. {8}, {T}: Discover 10.",
		}},
		Static: []game.StaticAbility{
			effects.GrantToAttached("reach"),
			effects.GrantAbilitiesToAttached(swashbucklersWhipGrant),
		},
		Activated: []effects.ActivatedAbility{
			effects.EquipAbility("{1}"),
		},
	})
}

const swashbucklersWhipGrant = "swashbucklers-whip/tap-and-discover"
