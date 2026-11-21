package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Furystoke Giant — Creature — Giant Warrior {3}{R}{R}, 3/3:
//
//	"When this creature enters, other creatures you control gain
//	 "{T}: This creature deals 2 damage to any target" until end of turn.
//	 Persist"
//
// The creatures are the ones you control as the trigger resolves
// (CR 611.2c), each given the ability as an ADR 0093 duration grant; a
// creature still under summoning sickness can't use it (CR 302.6). A
// persisted Giant hands the ability out again. Persist is
// PrintedKeywords (#2075).
//
// No simplification.
const furystokeGiantGrant = "furystoke-giant/tap-two-damage"

func init() {
	effects.Register(effects.Spec{
		OracleID:        "0cc62807-aeb4-4dfb-b001-4e40bf703497",
		Name:            "Furystoke Giant",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordPersist},
		Grants: []effects.AbilityGrant{{
			Key: furystokeGiantGrant,
			Activated: []effects.ActivatedAbility{{
				Label:   "{T}: This creature deals 2 damage to any target.",
				Cost:    effects.TapCost(),
				Targets: effects.TargetAny(),
				Effect:  effects.DealDamageToTheTarget(2),
			}},
			Text: "{T}: This creature deals 2 damage to any target.",
		}},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Furystoke Giant — other creatures you control gain \"{T}: 2 damage to any target\" until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					return effects.GrantAbilitiesFor{
						Match: effects.And(effects.Creature(), effects.YouControl(), effects.OtherThan(item.SourceCardID)),
						Keys:  []string{furystokeGiantGrant},
						Label: "Furystoke Giant — {T}: 2 damage to any target until end of turn",
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
