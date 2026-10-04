package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "0cc62807-aeb4-4dfb-b001-4e40bf703497",
		Name:            "Furystoke Giant",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordPersist},
		Grants: []AbilityGrant{{
			Key: furystokeGiantGrant,
			Activated: []ActivatedAbility{{
				Label:   "{T}: This creature deals 2 damage to any target.",
				Cost:    TapCost(),
				Targets: TargetAny(),
				Effect:  DealDamageToTheTarget(2),
			}},
			Text: "{T}: This creature deals 2 damage to any target.",
		}},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Furystoke Giant — other creatures you control gain \"{T}: 2 damage to any target\" until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					return GrantAbilitiesFor{
						Match: And(Creature(), YouControl(), OtherThan(item.SourceCardID)),
						Keys:  []string{furystokeGiantGrant},
						Label: "Furystoke Giant — {T}: 2 damage to any target until end of turn",
					}.Apply(NewContext(g, item))
				}),
		},
	})
}
