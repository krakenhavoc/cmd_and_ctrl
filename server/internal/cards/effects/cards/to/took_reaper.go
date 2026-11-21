package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Took Reaper — Creature — Halfling Peasant {1}{W}, 2/1:
//
//	"When this creature dies, the Ring tempts you."
//
// The tempt is the keyword action (ADR 0114, CR 701.54). The creature
// is in the graveyard by then, so it can't be chosen as the
// Ring-bearer; the trigger's controller is whoever controlled it as it
// died (CR 603.10a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bfd8daa9-029f-41aa-a947-cdfde3f4d6b6",
		Name:         "Took Reaper",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Took Reaper — the Ring tempts you", effects.Do(effects.TheRingTemptsYou{})),
		},
	})
}
