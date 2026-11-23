package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hell's Caretaker — Creature — Horror {3}{B}, 1/1:
//
//	"{T}, Sacrifice a creature: Return target creature card from your
//	 graveyard to the battlefield. Activate only during your upkeep."
//
// Priest of Fell Rites' reanimation with Magus of the Mirror's window.
// The target is chosen before the cost is paid (CR 602.2b follows CR
// 601.2c–h), so the creature sacrificed to pay is still on the
// battlefield when the target is picked and can't be the card it
// returns — the printed timing. The Caretaker may sacrifice itself.
// DuringYourUpkeep is the step AND your turn, so an opponent's upkeep
// is never a window.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2f0d797e-d897-453d-92b6-a90e1a548dc5",
		Name:         "Hell's Caretaker",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:     "{T}, Sacrifice a creature: Return target creature card from your graveyard to the battlefield. Activate only during your upkeep.",
			Cost:      effects.Plus(effects.TapCost(), effects.SacrificeACreature()),
			Condition: effects.DuringYourUpkeep(),
			Targets:   effects.TargetCardInGraveyard("target creature card from your graveyard", effects.Creature(), effects.YouOwn()),
			Effect:    effects.ReturnFirstLegalGraveyardTargetToBattlefield,
		}},
	})
}
