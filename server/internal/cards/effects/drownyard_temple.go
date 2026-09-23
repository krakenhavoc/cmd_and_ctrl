package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Drownyard Temple — Land (EDHREC rank 4267):
//
//	"{T}: Add {C}.
//	 {3}: Return this card from your graveyard to the battlefield
//	 tapped."
//
// The mana half is a plain colorless tap, the same shape as any
// basic. The graveyard half is a CR 602 activation reaching
// ActivatedAbilityShape.Zones (CR 113.6) exactly as Reassembling
// Skeleton's does one file over (#1221) — a mana cost with no tap
// component, since the land is not ON the battlefield to tap while
// this fires from its graveyard.
//
// "Enters tapped" is ReturnFromGraveyard.Tapped (#1284) — see
// reassembling_skeleton.go's doc comment for why that is a distinct
// primitive field rather than an OnETB tap.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c30f9be4-c274-4ad0-b5d7-7d3421aa4277",
		Name:         "Drownyard Temple",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []ActivatedAbility{{
			Label: "{3}: Return this card from your graveyard to the battlefield tapped.",
			Cost:  ManaCost("{3}"),
			Zones: []game.ZoneKind{game.ZoneGraveyard},
			Effect: func(g *game.Game, item *game.StackItem) error {
				return returnThisFromGraveyardTapped(g, item)
			},
		}},
	})
}
