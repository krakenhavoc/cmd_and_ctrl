package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sudden Spoiling — Instant {1}{B}{B}:
//
//	"Split second (As long as this spell is on the stack, players
//	 can't cast spells or activate abilities that aren't mana
//	 abilities.)
//	 Until end of turn, creatures target player controls lose all
//	 abilities and have base power and toughness 0/2."
//
// A combat trick for the whole of one player's board, behind split
// second (#1519) so they can't pump, sacrifice or flicker in response.
//
// One continuous effect in two layers — one data record (ADR 0041
// phase 3), one timestamp — pinned to one CR 611.2c set: the
// creatures that player controls as this resolves; one cast
// afterwards is untouched, and one that blinks is a new object
// (CR 400.7):
//
//   - layer 6, "lose all abilities" (CR 613.1f) — RemovesAbilities,
//     so catalogued triggered, activated, static and mana abilities
//     all go quiet as well as the keyword badges (the LoseAllAbilities
//     shape, pinned to the set rather than to an Aura's host);
//   - layer 7b, "base power and toughness 0/2" — SETS the value, so
//     an anthem's +1/+1 still applies on top (a 1/3), which is the
//     printed interaction and the reason this is not a 7c shrink.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "dce202c7-fe8e-462a-858e-7a5a69bd5b6b",
		Name:            "Sudden Spoiling",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordSplitSecond},
		Targets:         TargetPlayer("target player"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetPlayer {
				return nil
			}
			return untilEndOfTurn(ctx, uuid.Nil, And(Creature(), ControlledBy(item.Targets[0].ID)),
				"Sudden Spoiling — lose all abilities and have base 0/2",
				game.LoseAllAbilitiesMod(), game.SetBasePowerMod(0), game.SetBaseToughnessMod(2))
		},
	})
}
