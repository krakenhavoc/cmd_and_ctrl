package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Subcommander T'Pol — Legendary Creature — Vulcan Scientist {2}{U}, 3/1
// (Star Trek):
//
//	Flash
//	When Subcommander T'Pol enters, choose one —
//	• Counter target activated or triggered ability.
//	• Tap up to one target artifact or creature. Put a stun counter on
//	  it.
//
// One modal enters trigger (TriggeredAbility.Modes, #764). The first bullet
// is Sublime Epiphany's ability bullet (AbilityOnStack + CounterTheModesTarget,
// #1211); a mana ability never uses the stack, so it is never a target. The
// second takes zero or one target, so the bullet may be chosen with no
// artifact or creature on the battlefield and then does nothing; a chosen
// target is tapped and gets a stun counter (CR 122.1d).
//
// No simplification.
func init() {
	trigger := WhenThisEnters("Subcommander T'Pol — choose one",
		func(*game.Game, *game.StackItem) error { return nil })
	trigger.Modes = ChooseOne(
		ModeDoing("Counter target activated or triggered ability.",
			AbilityOnStack("target activated or triggered ability"),
			CounterTheModesTarget),
		ModeDoing("Tap up to one target artifact or creature. Put a stun counter on it.",
			TargetPermanent("up to one target artifact or creature", Or(Artifact(), Creature())).WithCount(0, 1),
			trekB9TapAndStunTheModesTarget),
	)
	Register(Spec{
		OracleID:        "e6973668-4ea7-4920-82a5-73abfc1c9fd5",
		Name:            "Subcommander T'Pol",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered:       []game.TriggeredAbility{trigger},
	})
}
