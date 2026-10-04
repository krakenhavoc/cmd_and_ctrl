package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rohirrim Lancer — Creature — Human Knight {R}, 1/1:
//
//	"Menace
//	 When this creature dies, the Ring tempts you."
//
// The tempt is the keyword action (ADR 0114, CR 701.54). The creature
// is in the graveyard by then, so it can't be chosen as the
// Ring-bearer; the trigger's controller is whoever controlled it as it
// died (CR 603.10a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "504fac88-1674-4157-becb-0204ab8844bf",
		Name:            "Rohirrim Lancer",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Rohirrim Lancer — the Ring tempts you", Do(TheRingTemptsYou{})),
		},
	})
}
