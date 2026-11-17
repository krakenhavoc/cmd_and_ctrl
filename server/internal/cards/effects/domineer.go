package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Domineer — Enchantment — Aura {1}{U}{U} (Mirrodin, #5606):
//
//	"Enchant artifact creature
//	 You control enchanted artifact creature."
//
// Mind Control on an artifact-creature enchant clause. The control
// change is the Aura's layer 2 continuous effect (CR 613.1b), not a
// one-shot, so it ends when the Aura leaves. The enchant clause is
// re-judged as a state-based action (CR 704.5m): a host that stops
// being an artifact creature (an animated artifact at end of turn, a
// creature that loses the artifact type) takes the Aura to the
// graveyard and goes back to its controller.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ca283839-864f-4252-8452-480e3ba29b20",
		Name:         "Domineer",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("enchant artifact creature", Artifact()),
		Static:       []game.StaticAbility{ControlAttachedBySource()},
	})
}
