package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Realm-Scorcher Hellkite — Creature — Dragon {4}{R}{R}, 4/6 (Wilds of
// Eldraine, #4471):
//
//	"Bargain (You may sacrifice an artifact, enchantment, or token as you
//	 cast this spell.)
//	 Flying, haste
//	 When this creature enters, if it was bargained, add four mana in any
//	 combination of colors.
//	 {1}{R}: This creature deals 1 damage to any target."
//
// Bargain is the shared optional cost (g2Bargain), and the intervening if
// (CR 603.4) is Diminisher Witch's: it reads the permanent's own record of
// what was paid, so an unbargained Hellkite, a reanimated one and a token
// copy put nothing on the stack. The enters trigger is not a mana ability
// (it does not trigger from one, CR 605.1b), so it uses the stack and adds
// Cosmic Crucible's four independent any-colour slots as it resolves. The
// ping is Ballista Watcher's "deals 1 damage to any target" for {1}{R},
// with no tap, so it can be activated as often as it is paid for.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "50ace222-57f1-477f-9e06-155d7b403512",
		Name:            "Realm-Scorcher Hellkite",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "haste"},
		OptionalCosts:   []game.AdditionalCost{effects.G2Bargain()},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, effects.AllOf(effects.Self, effects.DiminisherWitchWasBargained),
				"Realm-Scorcher Hellkite — add four mana in any combination of colors",
				effects.Do(effects.AddMana{Produced: effects.AnyCombinationOfColors(4)})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{R}: This creature deals 1 damage to any target.",
			Cost:    effects.ManaCost("{1}{R}"),
			Targets: effects.TargetAny(),
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 1)),
			Effect:  effects.SourceDealsDamageToEachLegalTarget(1),
		}},
	})
}
