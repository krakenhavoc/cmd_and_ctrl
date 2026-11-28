package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Equipoise — Enchantment {2}{W} (Visions, S82 OLD, #5631):
//
//	"At the beginning of your upkeep, for each land target player controls in
//	 excess of the number you control, choose a land that player controls,
//	 then the chosen permanents phase out. Repeat this process for artifacts
//	 and creatures."
//
// Three passes in the printed order (lands, artifacts, creatures), each
// counting when it is reached. The ability's controller makes the choices; a
// pass where every permanent is chosen anyway asks nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "aab905ef-b945-4523-b26f-fb5e6932bb56",
		Name:         "Equipoise",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.AtYourUpkeep("Equipoise — for each land target player controls in excess of the number you control, choose a land; repeat for artifacts and creatures", effects.S82OldB788EquipoiseUpkeep),
				effects.TargetPlayer("target player")),
		},
	})
}
