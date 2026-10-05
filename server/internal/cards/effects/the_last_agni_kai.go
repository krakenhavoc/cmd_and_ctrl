package effects

import (
	"strings"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Last Agni Kai — Instant {1}{R}:
//
//	"Target creature you control fights target creature an opponent
//	 controls. If the creature the opponent controls is dealt excess
//	 damage this way, add that much {R}.
//	 Until end of turn, you don't lose unspent red mana as steps and
//	 phases end."
//
// Excess is CR 120.4a's: what landed, less the damage the creature still
// needed to die, read after the fight. The last line is a granted
// player-level statement with an until-end-of-turn duration
// (game.GrantKeepManaForEffect, #2166), so it covers red mana from any
// source for the rest of the turn, not only the mana this spell adds.
func init() {
	Register(Spec{
		OracleID:     "fd1998c0-c18e-4310-82de-b6320be5fe10",
		Name:         "The Last Agni Kai",
		Completeness: CompletenessFull,
		Targets: Clauses(
			TargetCreature("target creature you control", YouControl()),
			TargetCreature("target creature an opponent controls", OpponentControls()),
		),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			ctx.Game.GrantKeepManaForEffect(item.Controller, []string{"R"},
				"The Last Agni Kai — keep unspent red mana", item.SourceCardID, ctx.Game.UntilEndOfTurnDuration())
			mine, ok := ctx.ClauseTarget(0)
			if !ok || mine.Kind != game.TargetCard {
				return nil
			}
			theirs, ok := ctx.ClauseTarget(1)
			if !ok || theirs.Kind != game.TargetCard {
				return nil
			}
			victim, ok := ctx.Game.LookupCardForEffect(theirs.ID)
			if !ok {
				return nil
			}
			lethal := max(victim.CurrentToughness()-victim.DamageMarked, 0)
			cursor := b25LastEventSeq(ctx.Game)
			if err := b10Fight(ctx, mine.ID, theirs.ID); err != nil {
				return err
			}
			excess := b27DamageDealtToAfter(ctx.Game, mine.ID, theirs.ID, cursor) - lethal
			if excess <= 0 {
				return nil
			}
			return AddMana{Produced: strings.Repeat("{R}", excess)}.Apply(ctx)
		},
	})
}
