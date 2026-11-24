package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Noxious Dragon — Creature — Dragon {4}{B}{B}, 4/4 (Commander Legends):
//
//	"Flying
//	 When this creature dies, you may destroy target creature with mana value 3
//	 or less."
//
// An optional targeted dies trigger (the target is chosen as the trigger goes
// on the stack, "you may" asked as it resolves). Mana value is read through
// the engine (CR 202.3), so a token is 0 and an X creature is 0 on the
// battlefield.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "fe75b364-9e65-499d-9763-76844cf88cf7",
		Name:            "Noxious Dragon",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.Targeting(effects.WhenThisDies("Noxious Dragon — destroy target creature with mana value 3 or less", effects.DestroyFirstLegalCardTarget),
				effects.TargetCreature("target creature with mana value 3 or less", effects.ManaValueLE(3))),
				"Noxious Dragon — destroy target creature with mana value 3 or less?"),
		},
	})
}
