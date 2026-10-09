package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Avatar of Burgeoning Echoes — Creature — Avatar {G}{U}, 2/3 (Reality
// Fracture, tracker #2795):
//
//	"Landfall — Whenever a land you control enters, empower Jace 2.
//	 Planeswalkers you control have "[−10]: Put a +1/+1 counter on target
//	 creature for each land you control.""
//
// Empower Jace is the keyword action (ADR 0139). The granted row is an
// ADR 0093 bundle with a loyalty cost (ADR 0140): the walker pays the
// ten. The lands are counted as the ability resolves and are the
// activator's; a target that left is skipped (CR 608.2b).
//
// No simplification.
func init() {
	const grant = "avatar-of-burgeoning-echoes/counters"
	Register(Spec{
		OracleID:     "08e8a63c-9fcc-48cf-a9af-15cbaeec943f",
		Name:         "Avatar of Burgeoning Echoes",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Landfall("Avatar of Burgeoning Echoes — empower Jace 2", Do(EmpowerJace{N: 2})),
		},
		Grants: []AbilityGrant{{
			Key: grant,
			Activated: []ActivatedAbility{{
				Label:   "−10: Put a +1/+1 counter on target creature for each land you control.",
				Cost:    LoyaltyCost(-10),
				Targets: TargetCreature("target creature"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					lands := countControlled(g, item.Controller, func(c game.Card) bool { return c.IsLand() })
					for _, t := range ctx.LegalTargets() {
						if t.Kind != game.TargetCard {
							continue
						}
						return AddCounter{Target: t.ID, Kind: game.CounterPlusOne, N: lands}.Apply(ctx)
					}
					return nil
				},
			}},
			Text: "[−10]: Put a +1/+1 counter on target creature for each land you control.",
		}},
		Static: []game.StaticAbility{GrantAbilitiesToYourPlaneswalkers(grant)},
	})
}
