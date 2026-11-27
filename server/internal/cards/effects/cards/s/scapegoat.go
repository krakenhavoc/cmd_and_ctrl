package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Scapegoat — Instant {W} (Stronghold, S82 OLD b749, #5627):
//
//	"As an additional cost to cast this spell, sacrifice a creature.
//	 Return any number of target creatures you control to their owner's hand."
//
// The sacrifice is an additional cost (CR 118.8), paid after the targets are
// chosen (CR 601.2c, 601.2h): a creature that was both sacrificed and chosen
// is no longer a legal target as the spell resolves (CR 608.2b) and stays in
// the graveyard. Zero targets is legal (CR 115.4); each target still legal
// goes back to its owner's hand.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "04eb4fc0-0cc5-44f1-af5a-44579125d5c2",
		Name:           "Scapegoat",
		Completeness:   effects.CompletenessFull,
		Purpose:        game.Purpose{Answers: game.AnswerProtect},
		AdditionalCost: effects.SacrificeCost("a creature", effects.Creature()),
		Targets:        effects.TargetCreature("any number of target creatures you control", effects.YouControl()).WithCount(0, 0),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.NeoB20BounceTargetedPermanents(ctx)
		},
	})
}
