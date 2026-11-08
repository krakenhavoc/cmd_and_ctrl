package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "50ace222-57f1-477f-9e06-155d7b403512",
		Name:            "Realm-Scorcher Hellkite",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "haste"},
		OptionalCosts:   []game.AdditionalCost{g2Bargain()},
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, AllOf(Self, diminisherWitchWasBargained),
				"Realm-Scorcher Hellkite — add four mana in any combination of colors",
				Do(AddMana{Produced: AnyCombinationOfColors(4)})),
		},
		Activated: []ActivatedAbility{{
			Label:   "{1}{R}: This creature deals 1 damage to any target.",
			Cost:    ManaCost("{1}{R}"),
			Targets: TargetAny(),
			Purpose: ForTargets(DamageToTarget(0, 1)),
			Effect:  sourceDealsDamageToEachLegalTarget(1),
		}},
	})
}
