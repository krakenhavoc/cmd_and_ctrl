package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Battlefly Swarm — Creature — Phyrexian Insect {B}, 1/1 (Dominaria United, S82):
//
//	"Flying
//	 {B}: This creature gains deathtouch until end of turn."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "1396e7f1-ac91-4778-9f71-493c7a44bd51",
		Name:            "Battlefly Swarm",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{B}: This creature gains deathtouch until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.ManaCost("{B}"),
			Effect:  effects.ThisGainsKeywordUntilEndOfTurn("deathtouch", "Battlefly Swarm — deathtouch until end of turn"),
		}},
	})
}
