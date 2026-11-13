package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Angler Turtle — Creature — Turtle {5}{U}{U}, 5/7 (M19 Gift Pack,
// #5524):
//
//	"Hexproof
//	 Creatures your opponents control attack each combat if able."
//
// Grand Melee's attack half (AttacksEachCombatWhere, CR 508.1d), scoped
// to creatures a player other than the Turtle's controller controls,
// read at every layer pass, so a creature an opponent casts or gains
// control of later is reached too. The requirement is the plain
// "attacks each combat if able": at a multiplayer table it is obeyed by
// attacking any player, planeswalker or battle the creature can attack,
// not only the Turtle's controller. A creature that can't attack, or
// could only by paying a cost, is not made to (CR 508.1d). The
// requirement belongs to the Turtle, so it is attributed to it in a
// refusal and ends when the Turtle leaves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "a09da13e-f8a3-410e-941a-9912d2cfe7de",
		Name:            "Angler Turtle",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"hexproof"},
		Static: []game.StaticAbility{
			AttacksEachCombatWhere(func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.Controller != source.Controller
			}),
		},
	})
}
