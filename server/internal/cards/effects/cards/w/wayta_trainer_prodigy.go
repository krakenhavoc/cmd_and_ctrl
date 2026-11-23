package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wayta, Trainer Prodigy — Legendary Creature — Human Warrior {R}{G}{W}, 1/5
// (Lost Caverns of Ixalan Commander):
//
//	"Haste
//	 {2}{G}, {T}: Target creature you control fights another target
//	 creature. This ability costs {2} less to activate if it targets two
//	 creatures you control.
//	 If a creature you control being dealt damage causes a triggered ability
//	 of a permanent you control to trigger, that ability triggers an
//	 additional time."
//
// The fight is Ulvenwald Tracker's (CR 701.14). The discount is a cost clause
// of the ability itself (CR 601.2f) that reads the announced targets, and
// applies only when both are creatures its controller controls. The third
// ability is a trigger doubler (CR 603.2d) whose cause is damage dealt to one
// of your creatures, Wayta included, and whose ability must be one of your
// permanents'; an enrage creature's trigger therefore happens twice.
func init() {
	discount := effects.CostsLess(2, "This ability costs {2} less to activate if it targets two creatures you control.",
		effects.Cmdr2B08ItTargetsTwoCreaturesYouControl())
	discount.ReadsTargets = true
	effects.Register(effects.Spec{
		OracleID:        "e507df33-d150-4b21-8ddb-90799f412dd8",
		Name:            "Wayta, Trainer Prodigy",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		TriggerDoublers: []game.TriggerDoubler{effects.Cmdr2B08CreatureYouControlDealtDamageDoubler()},
		Activated: []effects.ActivatedAbility{{
			Label: "{2}{G}, {T}: Target creature you control fights another target creature.",
			Cost:  effects.Plus(effects.ManaCost("{2}{G}"), effects.TapCost()),
			Targets: effects.Clauses(
				effects.TargetCreature("target creature you control", effects.YouControl()),
				effects.Distinct(effects.TargetCreature("another target creature")),
			),
			CostModifiers: []game.CostModifier{discount},
			Effect:        effects.SlotZeroFightsSlotOne,
		}},
	})
}
