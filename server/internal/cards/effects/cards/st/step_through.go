package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Step Through — Sorcery {3}{U}{U} (Modern Horizons 2, #5501):
//
//	"Return two target creatures to their owners' hands.
//	 Wizardcycling {2}"
//
// Exactly two targets are announced. They go back in one return (Whiplash
// Trap's), and a target that left in response is skipped while the other
// still returns (CR 608.2b). Wizardcycling is typecycling (CR 702.29e) for
// a card with the Wizard subtype.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c4fa89c7-f00d-42ef-bd67-407a288cd054",
		Name:         "Step Through",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("two target creatures").WithCount(2, 2),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			ctx.Game.BounceCardsToHandForEffect(effects.LegalTargetCardIDs(ctx))
			return nil
		},
		Activated: []effects.ActivatedAbility{
			effects.Typecycling("Wizardcycling", "{2}", "a Wizard card", effects.LrwB1CardOfType("Wizard")),
		},
	})
}
