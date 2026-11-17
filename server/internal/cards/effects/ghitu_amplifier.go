package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ghitu Amplifier — Creature — Human Wizard {1}{R}, 1/2 (Dominaria United):
//
//	"Kicker {2}{U}
//	 When this creature enters, if it was kicked, return target creature an
//	 opponent controls to its owner's hand.
//	 Whenever you cast an instant or sorcery spell, this creature gets +2/+0
//	 until end of turn."
//
// Two separate printed abilities, two registered triggers: Kor Sanctifiers'
// kicked enters trigger (an intervening "if", CR 603.4, so an unkicked
// Amplifier asks for no target) with a bounce, and Frolicking Familiar's
// cast-an-instant-or-sorcery pump.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "a965898b-9c68-46cb-82a5-a1e11397f68f",
		Name:          "Ghitu Amplifier",
		Completeness:  CompletenessFull,
		OptionalCosts: []game.AdditionalCost{Kicker("{2}{U}")},
		Triggered: []game.TriggeredAbility{
			Targeting(
				On(game.EventETB, AllOf(Self, ThisKickedAtLeast(1)),
					"Ghitu Amplifier — kicked, return target creature an opponent controls to its owner's hand",
					bounceTheTargetToHand),
				TargetCreature("target creature an opponent controls", OpponentControls())),
			WheneverYouCast(Or(Instant(), Sorcery()),
				"Ghitu Amplifier — gets +2/+0 until end of turn",
				thisGetsUntilEndOfTurn(2, 0, "Ghitu Amplifier — +2/+0")),
		},
	})
}
