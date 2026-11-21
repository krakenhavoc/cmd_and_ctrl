package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mirrormere Guardian — Creature — Dwarf Soldier {2}{G}, 4/2:
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
		OracleID:     "97c188c4-3f27-428e-a4f6-4213f44898fc",
		Name:         "Mirrormere Guardian",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Mirrormere Guardian — the Ring tempts you", effects.Do(effects.TheRingTemptsYou{})),
		},
	})
}
