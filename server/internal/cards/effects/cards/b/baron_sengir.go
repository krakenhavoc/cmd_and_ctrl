package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Baron Sengir — Legendary Creature — Vampire Noble {5}{B}{B}{B}, 5/5 (Homelands,
// S82 OLD b822, #5635):
//
//	"Flying
//	Whenever a creature dealt damage by Baron Sengir this turn dies, put a +2/+2
//	counter on Baron Sengir.
//	{T}: Regenerate another target Vampire."
//
// Sengir Vampire's trigger with a +2/+2 counter (a power/toughness counter,
// CR 122.1a), and Rushwood Herbalist's regeneration shield on another Vampire.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "fc728cd7-2b5b-45f3-896a-a45648bcb61d",
		Name:            "Baron Sengir",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.OldB822DamagedCreatureDies("Baron Sengir — put a +2/+2 counter on it", "+2/+2"),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Regenerate another target Vampire.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    effects.TapCost(),
			Targets: effects.Another(effects.TargetCreature("another target Vampire", effects.HasSubtype("Vampire"))),
			Effect:  effects.RegenerateTheTargetPermanent,
		}},
	})
}
