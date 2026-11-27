package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Cromat — Legendary Creature — Illusion {W}{U}{B}{R}{G}, 5/5 (Apocalypse,
// #5614):
//
//	"{W}{B}: Destroy target creature blocking or blocked by Cromat.
//	 {U}{R}: Cromat gains flying until end of turn.
//	 {B}{G}: Regenerate Cromat.
//	 {R}{W}: Cromat gets +1/+1 until end of turn.
//	 {G}{U}: Put Cromat on top of its owner's library."
//
// Five activated abilities, each a row on its own. The destroy target is a
// creature in combat with Cromat as the ability is put on the stack and as it
// resolves (CR 608.2b); regeneration is the engine's shield; the last puts
// Cromat on top of its owner's library, and does nothing once it has left or
// come back as a new object (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "376601b6-fe51-4e2d-8ec6-98f965d649a3",
		Name:         "Cromat",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{W}{B}: Destroy target creature blocking or blocked by Cromat.",
				Purpose: game.Purpose{Answers: game.AnswerRemove},
				Cost:    effects.ManaCost("{W}{B}"),
				Targets: effects.BlockingOrBlockedBySource(effects.TargetCreature("target creature blocking or blocked by Cromat")),
				Effect:  effects.S82OldB635DestroyTarget,
			},
			{
				Label:   "{U}{R}: Cromat gains flying until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerProtect},
				Cost:    effects.ManaCost("{U}{R}"),
				Effect:  effects.ThisGainsKeywordUntilEndOfTurn("flying", "Cromat — flying"),
			},
			effects.LrwB1RegenerateThisRow("{B}{G}"),
			{
				Label:   "{R}{W}: Cromat gets +1/+1 until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerPump},
				Cost:    effects.ManaCost("{R}{W}"),
				Effect:  effects.ThisGetsUntilEndOfTurn(1, 1, "Cromat — +1/+1 until end of turn"),
			},
			{
				Label:   "{G}{U}: Put Cromat on top of its owner's library.",
				Purpose: game.Purpose{Answers: game.AnswerSurvive},
				Cost:    effects.ManaCost("{G}{U}"),
				Effect:  effects.PutThisPermanentOnTopOfOwnersLibrary,
			},
		},
	})
}
