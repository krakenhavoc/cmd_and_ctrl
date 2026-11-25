package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Bond of Insight — Sorcery {3}{U} (War of the Spark, #5520):
//
//	"Each player mills four cards. Return up to two instant and/or sorcery
//	 cards from your graveyard to your hand. Exile Bond of Insight."
//
// Every library is milled in turn order, one prompt at a time, and the return
// follows once the last mill has ended, so a card the mills put into your
// graveyard is a candidate (CR 608.2c). The two cards are chosen as the clause
// resolves, not targeted. The spell exiles itself, so it is never a candidate
// for its own return.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3a877718-ff90-4d6e-9f95-47b06adad7f8",
		Name:         "Bond of Insight",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{SelfMillTutor: 1},
		OnResolve:    effects.S82OldB170BondOfInsight,
	})
}
