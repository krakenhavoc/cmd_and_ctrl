package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Necropanther — Creature — Cat Nightmare {1}{W}{B}, 3/3 (Ikoria: Lair of
// Behemoths, #5512):
//
//	"Mutate {2}{W/B}{W/B} (If you cast this spell for its mutate cost, put it
//	 over or under target non-Human creature you own. They mutate into the
//	 creature on top plus all abilities from under it.)
//	 Whenever this creature mutates, return target creature card with mana value
//	 3 or less from your graveyard to the battlefield."
//
// Mutate is the engine's (ADR 0223). The target is chosen as the trigger goes
// on the stack and re-checked as it resolves (CR 608.2b), with its mana value
// read in the graveyard.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "7df82bf8-b232-48e1-858b-1571c48b6f46",
		Name:             "Necropanther",
		Completeness:     effects.CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{effects.Mutate("{2}{W/B}{W/B}")},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.On(game.EventMutate, effects.ThisCreatureMutates,
				"Necropanther — return target creature card with mana value 3 or less from your graveyard to the battlefield", effects.ReturnFirstLegalGraveyardTargetToBattlefield),
				effects.TargetCardInGraveyard("target creature card with mana value 3 or less from your graveyard", effects.YouOwn(), effects.Creature(), effects.ManaValueLE(3))),
		},
	})
}
