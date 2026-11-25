package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Drag to the Underworld — Instant {2}{B}{B} (Theros Beyond Death, #5514):
//
//	"This spell costs {X} less to cast, where X is your devotion to black.
//	 Destroy target creature."
//
// The reduction is the spell's own (a self cost modifier), counted as it is
// priced: each {B} in the mana costs of permanents the caster controls
// (CR 700.5), read through the black of the card's own text. It lowers the
// generic part only (CR 601.2f), so the {B}{B} is always paid.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "746dc1d8-f245-47be-abd5-50ca6ac0d1af",
		Name:         "Drag to the Underworld",
		Completeness: effects.CompletenessFull,
		SelfCostModifiers: []game.CostModifier{
			effects.CostsLessEach(func(q game.CostQuery) int {
				return effects.DevotionTo(q.Game, q.Controller, q.Source.TextColor("B"))
			}, "This spell costs {X} less to cast, where X is your devotion to black."),
		},
		Targets:   effects.TargetCreature("target creature"),
		Purpose:   game.Purpose{Answers: game.AnswerRemove},
		OnResolve: effects.DestroyTheTargetPermanent,
	})
}
