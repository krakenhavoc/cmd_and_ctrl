package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fetid Imp — Creature — Imp {1}{B}, 1/2 (Magic Origins, #5549):
//
//	"Flying
//	 {B}: This creature gains deathtouch until end of turn. (Any amount of damage it deals to a creature is enough to destroy it.)"
//
// Battlefly Swarm's shape: flying and a self deathtouch grant until end
// of turn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "cdb68656-633d-436e-ad56-0fdb4814bd19",
		Name:            "Fetid Imp",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{B}: This creature gains deathtouch until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.ManaCost("{B}"),
			Effect:  effects.ThisGainsKeywordUntilEndOfTurn("deathtouch", "Fetid Imp — deathtouch until end of turn"),
		}},
	})
}
