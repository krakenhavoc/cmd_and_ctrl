package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:      "a965898b-9c68-46cb-82a5-a1e11397f68f",
		Name:          "Ghitu Amplifier",
		Completeness:  effects.CompletenessFull,
		OptionalCosts: []game.AdditionalCost{effects.Kicker("{2}{U}")},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.On(game.EventETB, effects.AllOf(effects.Self, effects.ThisKickedAtLeast(1)),
					"Ghitu Amplifier — kicked, return target creature an opponent controls to its owner's hand",
					effects.BounceTheTargetToHand),
				effects.TargetCreature("target creature an opponent controls", effects.OpponentControls())),
			effects.WheneverYouCast(effects.Or(effects.Instant(), effects.Sorcery()),
				"Ghitu Amplifier — gets +2/+0 until end of turn",
				effects.ThisGetsUntilEndOfTurn(2, 0, "Ghitu Amplifier — +2/+0")),
		},
	})
}
