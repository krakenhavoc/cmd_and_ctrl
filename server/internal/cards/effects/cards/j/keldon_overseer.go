package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Keldon Overseer — Creature — Human Warrior {2}{R}, 3/1 (Dominaria):
//
//	"Kicker {3}{R} (You may pay an additional {3}{R} as you cast this spell.)
//	 Haste
//	 When this creature enters, if it was kicked, gain control of target
//	 creature until end of turn. Untap that creature. It gains haste until end of
//	 turn."
//
// Act of Treason as a kicked enters trigger: an intervening "if" read off the
// cast record (CR 603.4), so an Overseer that was not cast was not kicked and
// steals nothing. The target is chosen as the trigger goes on the stack; the
// control, the untap and the haste each last as printed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "1b1dc78b-0102-49d6-8e54-89bc18a1628f",
		Name:            "Keldon Overseer",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		OptionalCosts:   []game.AdditionalCost{effects.Kicker("{3}{R}")},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.On(game.EventETB, effects.AllOf(effects.Self, effects.ThisKickedAtLeast(1)),
					"Keldon Overseer — kicked, gain control of target creature until end of turn; untap it; it gains haste",
					effects.OldB208ThreatenTarget("Keldon Overseer")),
				effects.TargetCreature("target creature")),
		},
	})
}
