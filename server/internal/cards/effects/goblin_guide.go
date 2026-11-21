package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Goblin Guide — Creature — Goblin Scout {R}, 2/2 (Mystery Booster 2):
//
//	"Haste
//	 Whenever this creature attacks, defending player reveals the top card
//	 of their library. If it's a land card, that player puts it into their
//	 hand."
//
// The defending player is the player attacked, or the controller of the
// planeswalker or battle attacked (CR 506.3d). It is that player's library
// that is revealed and their hand that gets the land; the Guide's controller
// gets nothing. A nonland card stays on top. A player who has left the
// game, or an empty library, does nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "51d9564b-44fc-4de1-9119-09d7b4089378",
		Name:            "Goblin Guide",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered: []game.TriggeredAbility{
			On(game.EventAttack, ThisAttacked,
				"Goblin Guide — defending player reveals the top card of their library; a land goes to their hand",
				func(g *game.Game, item *game.StackItem) error {
					return s82GapB22DefenderRevealsTopLandToHand(g, item, defendingPlayerOfAttack(g, NewContext(g, item).Trigger().Event))
				}),
		},
	})
}
