package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Smokestack — Artifact {4} (Mystery Booster 2):
//
//	"At the beginning of your upkeep, you may put a soot counter on this
//	 artifact.
//	 At the beginning of each player's upkeep, that player sacrifices a
//	 permanent of their choice for each soot counter on this artifact."
//
// Two printed abilities, two triggers. The sacrifice is the upkeep
// player's own choice, owed once per soot counter counted as the trigger
// resolves (CR 608.2h); a player with fewer permanents sacrifices what they
// have (CR 701.21a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "fda61506-ed79-47fd-badd-639185297913",
		Name:         "Smokestack",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep("Smokestack — you may put a soot counter on this artifact", smokestackMayAddSoot),
			AtEachUpkeep("Smokestack — that player sacrifices a permanent for each soot counter", smokestackSacrifice),
		},
	})
}

const smokestackSoot = "soot"

func smokestackMayAddSoot(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	return MayChoice{
		Question: "Smokestack — put a soot counter on it?",
		Then:     smokestackAddSoot,
		Carry:    []uuid.UUID{item.SourceCardID},
	}.Apply(ctx)
}

var smokestackAddSoot = MayYes("confirm/smokestack-add-soot", func(ctx *Context, r game.ConfirmAnswer) error {
	return AddCounter{Target: carryAt(r, 0), Kind: smokestackSoot, N: 1}.Apply(ctx)
})

func smokestackSacrifice(g *game.Game, item *game.StackItem) error {
	player := triggeringActor(item)
	src, ok := g.LookupCardForEffect(item.SourceCardID)
	if player == uuid.Nil || !ok || g.PlayerByIDForEffect(player) == nil {
		return nil
	}
	n := src.Counters[smokestackSoot]
	if n <= 0 {
		return nil
	}
	players := make([]uuid.UUID, n)
	for i := range players {
		players[i] = player
	}
	g.PlayersSacrificeForEffect(item.SourceCardID, players, nil, "Sacrifice a permanent")
	return nil
}
