package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Jaheira's Respite — Instant {4}{G} (Commander Legends: Battle for Baldur's
// Gate, #5948):
//
//	"Search your library for up to X basic land cards, where X is the number of
//	 creatures attacking you, put those cards onto the battlefield tapped, then
//	 shuffle.
//	 Prevent all combat damage that would be dealt this turn."
//
// X counts the creatures attacking you (not your planeswalkers) as the spell
// resolves (CR 608.2h), as Arachnogenesis does. The search is for up to X basic
// lands and may find fewer (CR 701.23b); with X = 0 nothing is searched and
// the library is not shuffled. The Fog shield (PreventAllCombatDamageThisTurn)
// waits for any prompt the search asks (CR 608.2c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1c19b04d-b307-4f70-b89e-af306374b9db",
		Name:         "Jaheira's Respite",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Lands: 1, Tutors: 1, Answers: game.AnswerPrevent},
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				n := effects.CreaturesAttackingPlayer(ctx.Game, ctx.Controller())
				if n <= 0 {
					return nil
				}
				return effects.SearchLibrary{
					Player:        ctx.Controller(),
					Predicate:     effects.IsBasicLand,
					Dest:          game.ZoneBattlefield,
					Limit:         n,
					Shuffle:       true,
					TappedOnEntry: true,
					Source:        ctx.Source(),
					Reason:        "Jaheira's Respite — up to X basic land cards, onto the battlefield tapped",
				}.Apply(ctx)
			}),
			effects.PreventAllCombatDamageThisTurn{Label: "Jaheira's Respite — prevent all combat damage that would be dealt this turn"},
		),
	})
}
