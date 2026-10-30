package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Giant-Sized Flying Ant — {3}{U} Creature — Insect 3/2 (Marvel Super
// Heroes, #4455):
//
//	"Flash
//	 Flying
//	 When this creature enters, choose one —
//	 • Tap target nonland permanent.
//	 • Untap target nonland permanent."
//
// Flash and flying are printed keywords. The enter trigger is a modal
// triggered ability (CR 700.2b: the mode is chosen as the trigger goes
// on the stack, and only a mode with a legal target can be chosen), the
// shape Tiller Engine uses; each bullet's target is any nonland
// permanent, yours or not.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "233e3b4b-faaf-451d-8a2e-e9f5a6ced671",
		Name:            "Giant-Sized Flying Ant",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash", "flying"},
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventETB},
			AppliesTo: Self,
			Key:       "Giant-Sized Flying Ant — tap or untap target nonland permanent",
			Effect:    func(*game.Game, *game.StackItem) error { return nil },
			Modes: ChooseOne(
				ModeDoing("Tap target nonland permanent.",
					TargetPermanent("target nonland permanent", Nonland()), TapTheModesTarget),
				ModeDoing("Untap target nonland permanent.",
					TargetPermanent("target nonland permanent", Nonland()), UntapTheModesTarget),
			),
		}},
	})
}
