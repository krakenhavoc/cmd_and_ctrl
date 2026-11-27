package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shizuko, Caller of Autumn — Legendary Creature — Snake Shaman {1}{G}{G}, 2/3 (Betrayers of
// Kamigawa, S82 OLD b595):
//
//	"At the beginning of each player's upkeep, that player adds {G}{G}{G}. Until end of turn,
//	 they don't lose this mana as steps and phases end."
//
// Eladamri's Vineyard's shape on the upkeep: the mana goes to the player whose upkeep it is
// (the event's actor), and carries the keep-until-end-of-turn rider (KeepManaUntilEndOfTurn).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c4984e18-9caf-4dd0-9c24-e60ccb27e2a0",
		Name:         "Shizuko, Caller of Autumn",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtEachUpkeep("Shizuko — that player adds {G}{G}{G}, kept until end of turn", effects.ShizukoMana),
		},
	})
}
