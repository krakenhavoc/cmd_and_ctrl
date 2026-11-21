package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	trigger := effects.WhenThisEnters("Subcommander T'Pol — choose one",
		func(*game.Game, *game.StackItem) error { return nil })
	trigger.Modes = effects.ChooseOne(
		effects.ModeDoing("Counter target activated or triggered ability.",
			effects.AbilityOnStack("target activated or triggered ability"),
			effects.CounterTheModesTarget),
		effects.ModeDoing("Tap up to one target artifact or creature. Put a stun counter on it.",
			effects.TargetPermanent("up to one target artifact or creature", effects.Or(effects.Artifact(), effects.Creature())).WithCount(0, 1),
			effects.TrekB9TapAndStunTheModesTarget),
	)
	effects.Register(effects.Spec{
		OracleID:        "e6973668-4ea7-4920-82a5-73abfc1c9fd5",
		Name:            "Subcommander T'Pol",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered:       []game.TriggeredAbility{trigger},
	})
}
