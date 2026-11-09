package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// The Western Cloud — Plane — Gobakhan (#4870):
//
//	"Prevent all damage that would be dealt to creatures and planeswalkers
//	 you control.
//	 Whenever chaos ensues, create three tapped Treasure tokens. They each
//	 deal 1 damage to each creature and each planeswalker."
//
// "You" is the planar controller (CR 901.6), re-read as control passes. The
// Treasures, not the plane, deal the damage (the ruling), so each is a damage
// source of its own and the prevention still applies to the planar
// controller's permanents.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4abae901-eb47-4318-bf64-322a893c2833",
		Name:         "The Western Cloud",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			PreventAllDamageDealtTo(PreventionStatic{
				To:    ToCreaturesAndPlaneswalkersYouControl,
				Label: "The Western Cloud — prevent all damage that would be dealt to creatures and planeswalkers you control",
			}),
		},
		Triggered: []game.TriggeredAbility{
			WheneverChaosEnsues("The Western Cloud — create three tapped Treasure tokens; they each deal 1 damage to each creature and each planeswalker",
				b6TreasuresPingEverything),
		},
	})
}
