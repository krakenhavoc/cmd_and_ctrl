package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Avatar of Hope — Creature — Avatar {6}{W}{W}, 4/9 (Eighth Edition, S82 OLD
// batch 609, #5607):
//
//	"If you have 3 or less life, this spell costs {6} less to cast.
//	 Flying
//	 This creature can block any number of creatures."
//
// A self cost modifier read as the spell is cast (CR 601.2f): {6} off the
// generic part when the caster's life total is 3 or less (Shadow of
// Mortality's seam). Block any number is Wall of Glare's CanBlockAnyNumber.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "e2f22cef-9494-4356-aa5c-619aecd7e74d",
		Name:            "Avatar of Hope",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		SelfCostModifiers: []game.CostModifier{
			effects.CostsLessEach(effects.OldB609LifeThreeOrLessSix, "If you have 3 or less life, this spell costs {6} less to cast."),
		},
		Static: []game.StaticAbility{effects.CanBlockAnyNumber(effects.SelfOnly)},
	})
}
