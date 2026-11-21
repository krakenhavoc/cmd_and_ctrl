package cardsw

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Woebringer Demon — Creature — Demon {3}{B}{B}, 4/4 (Ravnica Remastered,
// #5948):
//
//	"Flying
//	 At the beginning of each player's upkeep, that player sacrifices a
//	 creature of their choice. If the player can't, sacrifice this creature."
//
// The player whose upkeep it is picks which creature to sacrifice, from the
// creatures they control that can be sacrificed (CR 701.21a, CR 101.2). When
// they have none to sacrifice the Demon itself is sacrificed, by its
// controller. The Demon may itself be the creature chosen on its
// controller's upkeep.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "21ea9501-9a0f-43a3-a2cc-e9648325c45e",
		Name:            "Woebringer Demon",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.AtEachUpkeep("Woebringer Demon — that player sacrifices a creature of their choice; if they can't, sacrifice this creature", woebringerDemonUpkeep),
		},
	})
}

func woebringerDemonUpkeep(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	player := ctx.Trigger().Event.Actor
	var candidates []uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Controller == player && c.IsCreature() && !g.CantBeSacrificedForEffect(c.InstanceID) {
			candidates = append(candidates, c.InstanceID)
		}
	}
	if len(candidates) == 0 {
		return effects.SacrificeThisIfStillOnBattlefield(g, item)
	}
	return effects.SacrificeChoice{
		Player:     player,
		Candidates: candidates,
		Question:   "Woebringer Demon — sacrifice a creature",
	}.Apply(ctx)
}
