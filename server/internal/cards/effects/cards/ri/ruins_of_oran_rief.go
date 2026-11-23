package cardsri

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ruins of Oran-Rief — Land (Modern Horizons 3 Commander, #5948):
//
//	"This land enters tapped.
//	 {T}: Add {C}.
//	 {T}: Put a +1/+1 counter on target colorless creature that entered this
//	 turn."
//
// Oran-Rief, the Vastwood's "entered this turn" is the per-turn tally's
// per-object entry cell (b06EnteredThisTurn), not summoning sickness. The
// counter goes on ONE target creature that is colourless as the ability is
// put on the stack and again as it resolves (CR 608.2b); it goes through
// AddCounter, so Doubling Season and Hardened Scales apply.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7140f396-1bfa-4b28-ba28-fa15eba74652",
		Name:         "Ruins of Oran-Rief",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SelfEntersTapped()},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Put a +1/+1 counter on target colorless creature that entered this turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.TapCost(),
			Targets: effects.TargetCreature("target colorless creature that entered this turn",
				func(g *game.Game, _ uuid.UUID, c game.Card) bool {
					return c.IsColorless() && effects.B06EnteredThisTurn(g, c.InstanceID)
				}),
			Effect: effects.PlusOneCounterOnTheTarget,
		}},
	})
}
