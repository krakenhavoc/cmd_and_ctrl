package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Akki War Paint — Enchantment — Aura {R} (Neon Dynasty, #5493):
//
//	"Enchant artifact or creature
//	 As long as enchanted permanent is a creature, it gets +2/+1."
//
// The enchant clause is the spec the CR 704.5m check reruns, so the Aura
// stays on an artifact that stops being a creature; the pump waits for the
// host to be a creature (Careful Cultivation's condition).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e01ab6d8-8419-436e-a8fb-10aa46ca3d05",
		Name:         "Akki War Paint",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("enchant artifact or creature", effects.Or(effects.Artifact(), effects.Creature())),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static:       []game.StaticAbility{effects.B43PumpAttachedWhen(effects.Cmdr2B19AttachedIsCreature, 2, 1)},
	})
}
