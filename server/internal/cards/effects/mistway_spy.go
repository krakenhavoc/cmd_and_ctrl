package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mistway Spy — Creature — Merfolk Detective {U}, 1/1 (Murders at Karlov
// Manor, #4469):
//
//	"Flying
//	 Disguise {1}{U} (You may cast this card face down for {3} as a 2/2
//	 creature with ward {2}. Turn it face up any time for its disguise
//	 cost.)
//	 When this creature is turned face up, until end of turn, whenever a
//	 creature you control deals combat damage to a player, investigate."
//
// Turning it face up creates a repeating delayed triggered ability (CR
// 603.7b) that lasts the turn: Surge to Victory's "whenever a creature you
// control deals combat damage to a player this turn"
// (socB14YourCreatureDealtCombatDamageToAPlayerCondition), one Clue for
// each creature's combat damage event. The creature is read as it dealt
// the damage, so it need not be the Spy, and the Spy leaving the
// battlefield afterwards does not end the ability.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "87ccb151-a948-4e5a-878b-0d46c921a2b7",
		Name:             "Mistway Spy",
		Completeness:     CompletenessFull,
		PrintedKeywords:  []string{"flying"},
		AlternativeCosts: []game.AlternativeCost{Disguise("{1}{U}")},
		Triggered: []game.TriggeredAbility{
			WhenThisIsTurnedFaceUp("Mistway Spy — until end of turn, whenever a creature you control deals combat damage to a player, investigate",
				func(g *game.Game, item *game.StackItem) error {
					return DelayedOnEvent{
						Label:     "Mistway Spy — investigate",
						On:        []game.EventKind{game.EventDealDamage},
						Condition: socB14YourCreatureDealtCombatDamageToAPlayerCondition,
						Body:      investigateForTheControllerBody,
						Repeats:   true,
					}.Apply(NewContext(g, item))
				}),
		},
	})
}

// investigateForTheControllerBody is a delayed trigger's "investigate":
// one Clue for the delayed ability's controller (CR 701.16a). An on-disk
// key: never renamed, never reused.
var investigateForTheControllerBody = game.SimpleDelayedBody("investigate/controller",
	func(g *game.Game, item *game.StackItem) error {
		return Investigate{Player: item.Controller, N: 1}.Apply(NewContext(g, item))
	})
