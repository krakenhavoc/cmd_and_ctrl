package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Secret of Bloodbending — Sorcery — Lesson {U}{U}{U}{U}:
//
//	"As an additional cost to cast this spell, you may waterbend {10}.
//	 You control target opponent during their next combat phase. If
//	 this spell's additional cost was paid, you control that player
//	 during their next turn instead. (You see all cards that player
//	 could see and make all decisions for them.)
//	 Exile Secret of Bloodbending."
//
// Two seams (#3999):
//
//   - The waterbend is an optional additional cost (OptionalWaterbend):
//     chosen, it adds {10} to the cast, and each untapped artifact or
//     creature the caster taps pays {1} of it (CR 701.67).
//   - The control is a combat-phase record (game.BeginCombatControlForEffect,
//     ADR 0167's 2026-10-19 amendment): it waits for the opponent's next
//     combat phase that begins, an additional one included, and the
//     caster plays that phase for them. Paid, it is instead the plain
//     turn-long control Mindslaver gives (no extra turn).
//
// The exile is the spell's last instruction, as Worst Fears' is, so a
// countered or fizzled Secret goes to the graveyard like any other
// spell.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "743fa824-18cb-4225-ba7b-5b0bbc3c8d41",
		Name:          "Secret of Bloodbending",
		Completeness:  CompletenessFull,
		OptionalCosts: []game.AdditionalCost{OptionalWaterbend("{10}")},
		Targets:       TargetPlayer("target opponent", Opponent()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if p := firstLegalTargetPlayer(ctx); p != uuid.Nil {
				if ctx.OptionalCostTimes(game.WaterbendKey) > 0 {
					controlPlayerNextTurn(ctx.Game, item, item.Controller, p, "Secret of Bloodbending")
				} else {
					ctx.Game.BeginCombatControlForEffect(game.CombatControl{
						Controller: item.Controller,
						Controlled: p,
						Source:     item.SourceCardID,
						SourceName: "Secret of Bloodbending",
					})
				}
			}
			return ctx.Game.ExileCardForEffect(item.SourceCardID)
		},
	})
}
