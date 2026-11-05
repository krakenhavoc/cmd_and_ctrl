package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Raphael, Tag Team Tough — {4}{R}{R} Legendary Creature — Mutant Ninja
// Turtle 5/6 (Teenage Mutant Ninja Turtles Commander, #5011):
//
//	"Menace
//	 Whenever Raphael deals combat damage to a player for the first time
//	 each turn, untap all attacking creatures. After this combat phase,
//	 there is an additional combat phase."
//
// "For the first time each turn" is the per-object trigger tally Hercules,
// Olympian Hero reads (b11TriggeredThisTurn): the first combat damage
// Raphael deals to a player in a turn triggers it, and every later hit
// that turn (double strike's second strike, the added combat's damage) does
// not. A Raphael that left and came back is a new object and may trigger
// again (CR 400.7). The body is Desert Were-Worm's: untap every attacking
// creature, then one added combat phase after this one, with no main phase.
//
// No simplification.
func init() {
	const label = "Raphael, Tag Team Tough — untap all attacking creatures; after this combat phase, an additional combat phase"
	Register(Spec{
		OracleID:        "5a7f09d3-e4a8-4987-a731-a68fc748277d",
		Name:            "Raphael, Tag Team Tough",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			On(game.EventDealDamage, func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
				return ThisDealtCombatDamageToAPlayer(ev, source, lki, g) && !b11TriggeredThisTurn(g, source.InstanceID, label)
			}, label, Do(UntapAttackingCreatures{}, ExtraCombatAfterThisPhase())),
		},
	})
}
