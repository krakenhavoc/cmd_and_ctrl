package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "e01ab6d8-8419-436e-a8fb-10aa46ca3d05",
		Name:         "Akki War Paint",
		Completeness: CompletenessFull,
		Targets:      TargetPermanent("enchant artifact or creature", Or(Artifact(), Creature())),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static:       []game.StaticAbility{b43PumpAttachedWhen(cmdr2B19AttachedIsCreature, 2, 1)},
	})
}
