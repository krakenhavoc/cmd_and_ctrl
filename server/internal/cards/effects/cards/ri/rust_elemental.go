package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rust Elemental — Artifact Creature — Elemental {4}, 4/4 (Mirrodin, #5606):
//
//	"Flying
//	 At the beginning of your upkeep, sacrifice another artifact. If you
//	 can't, tap this creature and you lose 4 life."
//
// Xathrid Demon's upkeep sacrifice and its fallback, over artifacts: the
// controller chooses which other artifact to give up; with none to sacrifice
// the Elemental taps (if it is still there) and its controller loses 4 life.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "21af6b08-1dd1-4e7f-be7b-cc8e70c578c7",
		Name:            "Rust Elemental",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Rust Elemental — sacrifice another artifact; if you can't, tap this creature and you lose 4 life", effects.OldB606RustElementalUpkeep),
		},
	})
}
