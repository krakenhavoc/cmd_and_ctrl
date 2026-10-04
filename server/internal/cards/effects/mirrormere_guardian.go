package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "97c188c4-3f27-428e-a4f6-4213f44898fc",
		Name:         "Mirrormere Guardian",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Mirrormere Guardian — the Ring tempts you", Do(TheRingTemptsYou{})),
		},
	})
}
