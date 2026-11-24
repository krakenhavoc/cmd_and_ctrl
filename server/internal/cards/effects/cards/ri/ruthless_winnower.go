package cardsri

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ruthless Winnower — Creature — Elf Rogue {3}{B}{B}, 4/4 (Kaldheim
// Commander, #5506):
//
//	"At the beginning of each player's upkeep, that player sacrifices a
//	 non-Elf creature of their choice."
//
// Woebringer Demon's upkeep edict: it fires on every player's upkeep, and that
// player picks among the non-Elf creatures they control (read through the
// layered view; a creature that can't be sacrificed is not offered). With
// none, nothing happens. A player with exactly one is not asked.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "61f68cf6-e083-4264-89b4-cc17749b2a8b",
		Name:         "Ruthless Winnower",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtEachUpkeep("Ruthless Winnower — that player sacrifices a non-Elf creature of their choice", ruthlessWinnowerUpkeep),
		},
	})
}

func ruthlessWinnowerUpkeep(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	player := ctx.Trigger().Event.Actor
	elf := ctx.TextSubtype("Elf")
	var candidates []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == player && c.IsCreature() && !c.HasSubtype(elf) && !g.CantBeSacrificedForEffect(c.InstanceID) {
			candidates = append(candidates, c.InstanceID)
		}
	}
	if len(candidates) == 0 {
		return nil
	}
	return effects.SacrificeChoice{
		Player:     player,
		Candidates: candidates,
		Question:   "Ruthless Winnower — sacrifice a non-Elf creature",
	}.Apply(ctx)
}
