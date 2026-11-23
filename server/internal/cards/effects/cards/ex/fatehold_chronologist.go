package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fatehold Chronologist // Peer Review — Creature — Bird Wizard {1}{W/U},
// 1/2 // Sorcery {2}{W/U} (preparation card, CR 722):
//
//	"Flying
//	 This creature enters prepared. (While it's prepared, you may cast a
//	 copy of its spell. Doing so unprepares it.)"
//
//	Peer Review — "Create a 2/2 colorless Wizard Soldier creature token
//	 named Cadet. Surveil 1."
//
// The surveil runs after the token is made, in printed order.
//
// No simplification.
func init() {
	const id = "1063822f-47d3-42e9-8a21-f62b12609fe1"
	effects.Register(effects.Spec{
		OracleID:        id,
		Name:            "Fatehold Chronologist",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Replacements:    []game.ReplacementEffect{effects.SelfEntersPrepared()},
	})
	effects.Register(effects.Spec{
		OracleID:     id + "#1",
		Name:         "Peer Review",
		Completeness: effects.CompletenessFull,
		OnResolve:    effects.PeerReviewResolve,
	})
}
