package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Rack — Artifact {1}:
//
//	"As this artifact enters, choose an opponent.
//	 At the beginning of the chosen player's upkeep, this artifact
//	 deals X damage to that player, where X is 3 minus the number of
//	 cards in their hand."
//
// Storm World's damage pointed at one stored seat. The choice is made
// as the artifact enters (ChoosePlayerAsEnters, the pool restricted to
// the controller's opponents) and stored on it, and the trigger reads
// that seat live on every upkeep, the way Sawhorn Nemesis's
// replacement does. Until the controller answers the prompt nobody is
// chosen and the trigger fires for no one. X is counted as the
// trigger resolves, and a hand of three or more takes nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3d873e1d-4fac-42c4-bb31-77e76099e1ef",
		Name:         "The Rack",
		Completeness: CompletenessFull,
		AsEnters:     ChoosePlayerAsEnters("The Rack", Opponents),
		Triggered: []game.TriggeredAbility{
			On(game.EventBeginUpkeep, theChosenPlayersTurn,
				"The Rack — deals damage to the chosen player equal to 3 minus their hand size",
				damageUpkeepPlayerByHand(func(hand int) int { return 3 - hand })),
		},
	})
}

// theChosenPlayersTurn — the event's actor (the upkeep's player) is the
// seat stored on the source by its as-enters choice.
func theChosenPlayersTurn(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	chosen := ChosenPlayerOf(g, source.InstanceID)
	return chosen != uuid.Nil && ev.Actor == chosen
}
