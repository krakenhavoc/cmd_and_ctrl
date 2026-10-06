package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Haze Frog — Creature — Frog {3}{G}{G}, 2/1:
//
//	"Flash
//	 When this creature enters, prevent all combat damage that other
//	 creatures would deal this turn."
//
// #2026: the Frog is pinned as the object its trigger came from (CR
// 400.7) and left out of the shield; every other creature's combat
// damage is prevented, read as it would be dealt (CR 609.7b). The
// rulings: creatures that weren't on the battlefield as the ability
// resolved are caught, and two Frogs that enter in one turn each stop
// the other's damage. A Frog that leaves and comes back is a new object,
// and its damage is prevented.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:        "d4cbf16d-6da8-4018-8398-8e263ef6c69e",
		Name:            "Haze Frog",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Haze Frog — prevent all combat damage other creatures would deal this turn",
				Do(combatShieldAgainstCreatures(game.DamageSourceFilter{}).OtherThanThis())),
		},
	})
}
