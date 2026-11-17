package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Geode Rager — Creature — Elemental {4}{R}{R}, 4/3 (Zendikar Rising
// Commander, #5509):
//
//	"First strike
//	 Landfall — Whenever a land you control enters, goad each creature
//	 target player controls. (Until your next turn, those creatures
//	 attack each combat if able and attack a player other than you if
//	 able.)"
//
// Landfall is an ability word (CR 207.2c): the trigger is Landfall with
// a "target player" clause, any player at the table, chosen as the
// trigger goes on the stack and checked again as it resolves (CR 603.3d,
// CR 608.2b). "Each creature target player controls" is read as the
// trigger resolves, so a creature that enters or changes control
// afterwards is not goaded (the ruling). The goad is the trigger
// controller's (GoadAllMatching, CR 701.15a), Komainu Battle Armor's
// body with the player taken from the target instead of the damage.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "3969d1d6-0aed-4fa8-aa09-d9be50576fb4",
		Name:            "Geode Rager",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Triggered: []game.TriggeredAbility{
			Targeting(
				Landfall("Geode Rager — goad each creature target player controls (landfall)", geodeRagerGoad),
				TargetPlayer("target player")),
		},
	})
}

// geodeRagerGoad goads every creature the target player controls as the
// trigger resolves. A target that left the game does nothing.
func geodeRagerGoad(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	player := TargetedPlayer(ctx)
	if player == uuid.Nil {
		return nil
	}
	return GoadAllMatching(ctx, func(_ *game.Game, c game.Card) bool {
		return c.Controller == player
	})
}
