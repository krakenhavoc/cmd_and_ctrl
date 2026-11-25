package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Boneyard Lurker — Creature — Nightmare Beast {2}{B}{G}, 4/4 (Ikoria: Lair
// of Behemoths, #5512):
//
//	"Mutate {2}{B/G}{B/G} (If you cast this spell for its mutate cost, put it
//	 over or under target non-Human creature you own. They mutate into the
//	 creature on top plus all abilities from under it.)
//	 Whenever this creature mutates, return target permanent card from your
//	 graveyard to your hand."
//
// Mutate is the engine's (ADR 0223). The target is chosen as the trigger goes
// on the stack (CR 603.3d) and re-checked as it resolves (CR 608.2b); a token
// is not a card, so none is ever a target.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "88aa5cd0-811a-4682-b2b0-2decbd13b332",
		Name:             "Boneyard Lurker",
		Completeness:     effects.CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{effects.Mutate("{2}{B/G}{B/G}")},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.On(game.EventMutate, effects.ThisCreatureMutates,
				"Boneyard Lurker — return target permanent card from your graveyard to your hand", effects.ReturnTargetCardToHand),
				effects.TargetCardInGraveyard("target permanent card from your graveyard", effects.YouOwn(), effects.Permanent())),
		},
	})
}
