package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pull from the Grave — Sorcery {2}{B} (Secrets of Strixhaven, #4456):
//
//	"Return up to two target creature cards from your graveyard to your
//	 hand. You gain 2 life."
//
// Macabre Waltz's returns (up to two, so a spell with no target still
// resolves), then the life. When every chosen target has left, the spell
// does nothing at all (CR 608.2b), the life included; with none chosen it
// gains the 2 life.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c5373026-8b43-42a7-aa69-a1b0f1619795",
		Name:         "Pull from the Grave",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCardInGraveyard("up to two target creature cards from your graveyard", effects.YouOwn(), effects.Creature()).WithCount(0, 2),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(effects.ReturnLegalGraveyardTargetsToHand),
			effects.Instruction(func(ctx *effects.Context) error { return effects.GainLife{Amount: 2}.Apply(ctx) }),
		),
	})
}
