package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Shenanigans — Sorcery {1}{R} (Modern Horizons, #5519):
//
//	"Destroy target artifact.
//	 Dredge 1 (If you would draw a card, you may mill a card instead. If you
//	 do, return this card from your graveyard to your hand.)"
//
// Skycrash's destroy and Darkblast's dredge replacement at 1.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1d9b94af-3bfa-4257-8347-675468f0b0e5",
		Name:         "Shenanigans",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetPermanent("target artifact", Artifact()),
		OnResolve:    destroyTheTargetPermanent,
		Replacements: []game.ReplacementEffect{Dredge(1)},
	})
}
