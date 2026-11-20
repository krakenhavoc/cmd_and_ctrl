package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Keepsake Gorgon — Creature — Gorgon {3}{B}{B}, 2/5 (Theros, #5561):
//
//	"Deathtouch
//	 {5}{B}{B}: Monstrosity 1. (If this creature isn't monstrous, put a +1/+1 counter on it and it becomes monstrous.)
//	 When this creature becomes monstrous, destroy target non-Gorgon creature an opponent controls."
//
// Arbor Colossus' shape: a mandatory targeted becomes-monstrous trigger.
// With no non-Gorgon creature across the table it is removed as it would
// go on the stack (CR 603.3d) and the Gorgon is simply monstrous.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "3aa58e5a-1b5e-4fb3-bf2d-9bad8126b6e6",
		Name:            "Keepsake Gorgon",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Activated:       []ActivatedAbility{Monstrosity(ManaCost("{5}{B}{B}"), 1)},
		Triggered:       []game.TriggeredAbility{keepsakeGorgonTrigger()},
	})
}

func keepsakeGorgonTrigger() game.TriggeredAbility {
	t := WhenBecomesMonstrous("Keepsake Gorgon — destroy target non-Gorgon creature an opponent controls", destroyChosenPermanent)
	t.Targets = TargetCreature("target non-Gorgon creature an opponent controls", Not(Subtype("Gorgon")), OpponentControls())
	return t
}
