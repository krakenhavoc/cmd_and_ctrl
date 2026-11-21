package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vona, Butcher of Magan — Legendary Creature — Vampire Knight {3}{W}{B},
// 4/4 (Commander, #5368):
//
//	"Vigilance, lifelink
//	 {T}, Pay 7 life: Destroy target nonland permanent. Activate only
//	 during your turn."
//
// {T} and the seven life are costs (CR 602.2b, 119.4): a player at six
// or less life can't pay. "Only during your turn" is an activation
// condition (DuringYourTurn), any step of it, not a sorcery-speed window.
// Destruction honours regeneration and indestructible as any destroy.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "b1c56e23-eae6-47a1-a1cd-3a2d01621768",
		Name:            "Vona, Butcher of Magan",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance", "lifelink"},
		Activated: []effects.ActivatedAbility{{
			Label:     "{T}, Pay 7 life: Destroy target nonland permanent. Activate only during your turn.",
			Purpose:   game.Purpose{Answers: game.AnswerRemove},
			Cost:      effects.Plus(effects.TapCost(), effects.PayLife(7)),
			Condition: effects.DuringYourTurn(),
			Targets:   effects.TargetPermanent("target nonland permanent", effects.Nonland()),
			Effect:    effects.DestroyChosenPermanent,
		}},
	})
}
