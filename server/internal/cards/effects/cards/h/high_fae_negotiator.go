package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// High Fae Negotiator — Creature — Faerie Warlock {3}{B}{B}, 3/5 (Wilds
// of Eldraine, #4471):
//
//	"Bargain (You may sacrifice an artifact, enchantment, or token as you
//	 cast this spell.)
//	 Flying
//	 When this creature enters, if it was bargained, each opponent loses
//	 3 life and you gain 3 life."
//
// Bargain is the shared optional cost (g2Bargain). The intervening if
// (CR 603.4) is Diminisher Witch's: it reads the permanent's own record
// of what was paid (ADR 0073 §5), so an unbargained Negotiator, a
// reanimated one and a token copy put nothing on the stack. The drain
// is Sneering Shadewriter's at 3: you gain 3 whatever the opponents
// lost.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "90ecd452-333a-43b0-be08-a61f3779b3bd",
		Name:            "High Fae Negotiator",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		OptionalCosts:   []game.AdditionalCost{effects.G2Bargain()},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, effects.AllOf(effects.Self, effects.DiminisherWitchWasBargained),
				"High Fae Negotiator — each opponent loses 3 life and you gain 3 life",
				effects.Do(effects.EachOpponentLosesLife{N: 3}, effects.GainLife{Amount: 3})),
		},
	})
}
