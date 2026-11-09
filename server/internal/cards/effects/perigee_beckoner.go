package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Perigee Beckoner — Creature — Horror {4}{B}, 4/5 (Edge of Eternities,
// #4461):
//
//	"When this creature enters, until end of turn, another target
//	 creature you control gets +2/+0 and gains "When this creature dies,
//	 return it to the battlefield tapped under its owner's control."
//	 Warp {1}{B}"
//
// Fake Your Own Death's duration grant on an enters trigger: one effect,
// one timestamp, +2/+0 and the dies trigger together, pinned to the target
// as it resolves (CR 611.2c). The dies trigger is the creature's own, fired
// from last-known information, and the creature that comes back is a new
// object (CR 400.7), so it does not return a second time. The target is
// "another" creature you control, so the Beckoner never targets itself,
// and with none the trigger is removed (CR 603.3d). Warp is the shared
// constructor.
//
// No simplification.
const perigeeBeckonerReturn = "perigee-beckoner/return"

func init() {
	enters := Targeting(
		WhenThisEnters("Perigee Beckoner — another target creature you control gets +2/+0 and returns when it dies",
			func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				target, ok := b16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				return GrantAbilitiesFor{
					Target: target,
					Keys:   []string{perigeeBeckonerReturn},
					Also:   []game.Mod{game.ModifyPTMod(2, 0)},
					Label:  "Perigee Beckoner — +2/+0 and it returns when it dies",
				}.Apply(ctx)
			}),
		Another(TargetCreature("another target creature you control", YouControl())))
	Register(Spec{
		OracleID:         "e0a35a44-e56d-434e-8043-43d5031b2e84",
		Name:             "Perigee Beckoner",
		Completeness:     CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{Warp("{1}{B}")},
		Grants: []AbilityGrant{{
			Key: perigeeBeckonerReturn,
			Triggered: []game.TriggeredAbility{
				WhenThisDies("Perigee Beckoner — return it tapped",
					returnThisCreatureFromGraveyard(nil, nil)),
			},
			Text: "When this creature dies, return it to the battlefield tapped under its owner's control.",
		}},
		Triggered: []game.TriggeredAbility{enters},
	})
}
