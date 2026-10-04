package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Relentless Rohirrim — Creature — Human Knight {3}{R}, 4/3:
//
//	"When this creature enters, the Ring tempts you."
//
// The tempt is the keyword action (ADR 0114, CR 701.54). The creature
// is on the battlefield as the trigger resolves, so it can be chosen as
// the Ring-bearer.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1212cc74-ff66-479e-8e1b-504446e1c6d9",
		Name:         "Relentless Rohirrim",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Relentless Rohirrim — the Ring tempts you", Do(TheRingTemptsYou{})),
		},
	})
}
