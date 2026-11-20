package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ravenous Gigantotherium — Creature — Beast {5}{G}{G}, 3/3 (Commander 2020,
// #5513):
//
//	"Devour 3 (As this creature enters, you may sacrifice any number of
//	 creatures. It enters with three times that many +1/+1 counters on it.)
//	 When this creature enters, it deals X damage divided as you choose among
//	 up to X target creatures, where X is its power. Each of those creatures
//	 deals damage equal to its power to this creature."
//
// Devour is the keyword (devour.go). The entry trigger is Polukranos, World
// Eater's monstrous bite on any creatures: X is the Gigantotherium's power
// as the trigger is put on the stack, devoured counters included, and the
// division is announced then (CR 601.2d); then each target still legal
// deals damage equal to its power to the Gigantotherium, simultaneously
// (polukranosBites).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "022624eb-c9d7-4110-bcc3-465c8d48fcd7",
		Name:         "Ravenous Gigantotherium",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{Devour("Ravenous Gigantotherium", 3)},
		Triggered:    []game.TriggeredAbility{ravenousGigantotheriumTrigger()},
	})
}

func ravenousGigantotheriumTrigger() game.TriggeredAbility {
	t := WhenThisEnters("Ravenous Gigantotherium — X damage divided among up to X target creatures; each deals damage equal to its power to it",
		polukranosBites)
	t.TargetsFrom = func(_ game.TriggerContext, source *game.Card, _ *game.Game) *game.TargetSpec {
		x := source.CurrentPower()
		if x <= 0 {
			return nil
		}
		return TargetCreature("up to X target creatures").WithCount(0, x).Dividing(Divide(x))
	}
	return t
}
