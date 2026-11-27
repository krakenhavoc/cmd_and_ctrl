package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hunter of Eyeblights — Creature — Elf Assassin {3}{B}{B}, 3/3 (Lorwyn, S82 OLD,
// #5593):
//
//	"When this creature enters, put a +1/+1 counter on target creature you don't
//	 control.
//	 {2}{B}, {T}: Destroy target creature with a counter on it."
//
// Bat Colony's counter on the clause's target (an opponent's creature, re-checked
// as the trigger resolves) and Glissa Sunslayer's "has a counter" clause for the
// destroy, which is offered only against a creature that carries one (any kind).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "34d5f249-df46-49d1-a692-da2348fbc51e",
		Name:         "Hunter of Eyeblights",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Hunter of Eyeblights — put a +1/+1 counter on target creature you don't control",
					effects.Cmdr2S46PlusOneCounterOnClauseTarget),
				effects.TargetCreature("target creature you don't control", effects.OpponentControls())),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{B}, {T}: Destroy target creature with a counter on it.",
			Cost:    effects.Plus(effects.ManaCost("{2}{B}"), effects.TapCost()),
			Targets: effects.TargetCreature("target creature with a counter on it", effects.B64HasAnyCounter()),
			Effect:  effects.DestroyFirstLegalTarget,
		}},
	})
}
