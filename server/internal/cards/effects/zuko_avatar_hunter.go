package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Zuko, Avatar Hunter — Legendary Creature — Human Noble {3}{R}{R}, 4/5
// (Avatar: The Last Airbender Eternal, #5075):
//
//	"Reach (This creature can block creatures with flying.)
//	 Whenever you cast a red spell, create a 2/2 red Soldier creature
//	 token."
//
// The cast trigger is Balefire Liege's: it fires on casting (CR 601.2i),
// so it resolves before the spell does, and the spell's colour is read
// as it is cast. The token is the set's plain 2/2 red Soldier.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "d445a584-98e7-4501-bb20-e76c9cafa757",
		Name:            "Zuko, Avatar Hunter",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Triggered: []game.TriggeredAbility{
			WheneverYouCast(OfColor("R"), "Zuko, Avatar Hunter — create a 2/2 red Soldier",
				Do(CreateToken{Template: TokenCard("2/2 red Soldier"), N: 1})),
		},
	})
}
