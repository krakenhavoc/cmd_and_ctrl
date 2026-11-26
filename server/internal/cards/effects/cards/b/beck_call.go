package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Beck // Call — split card, oracle 5b8472c8-a7e7-46f2-aecf-e954082a5ef5
// (Dragon's Maze, S82 OLD batch 402, #5564):
//
//	Beck — Sorcery {G}{U}: "Whenever a creature enters this turn, you may draw a
//	       card."
//	Call — Sorcery {4}{W}{U}: "Create four 1/1 white Bird creature tokens with
//	       flying."
//	       Fuse (You may cast one or both halves of this card from your hand.)
//
// Both halves are ADR 0034 faces (Beck under the bare oracle ID, Call under
// "#1"); the fused cast is the engine's (CR 702.102), Beck resolving first, so a
// fused Call's four Birds each trigger Beck's draw: they enter after the delayed
// trigger exists. Beck is Rite of Harmony's repeating delayed trigger (CR 603.7b)
// over any player's creature entering, with the "you may" asked of its controller
// as each trigger resolves (CR 603.5); a creature that entered before Beck
// resolved does not trigger it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5b8472c8-a7e7-46f2-aecf-e954082a5ef5",
		Name:         "Beck",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.DelayedOnEvent{
				Label:            "Beck — you may draw a card",
				On:               []game.EventKind{game.EventETB},
				Condition:        effects.S82OldB402AnyCreatureEntersCondition,
				OptionalQuestion: "Beck — draw a card?",
				Body:             effects.S82OldB402DrawACardBody,
				Repeats:          true,
			}.Apply(ctx)
		},
	})
	effects.Register(effects.Spec{
		OracleID:     "5b8472c8-a7e7-46f2-aecf-e954082a5ef5#1",
		Name:         "Call",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 4, CreatureTokens: 4},
		OnResolve:    effects.ResolveSteps(effects.CreateToken{Template: effects.TokenCard("1/1 white Bird with flying"), N: 4}),
	})
}
