package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:        "504fac88-1674-4157-becb-0204ab8844bf",
		Name:            "Rohirrim Lancer",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Rohirrim Lancer — the Ring tempts you", effects.Do(effects.TheRingTemptsYou{})),
		},
	})
}
