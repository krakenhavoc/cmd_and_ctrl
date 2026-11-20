package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Surge of Salvation — Instant {W}:
//
//	"You and permanents you control gain hexproof until end of turn.
//	 Prevent all damage that black and/or red sources would deal to
//	 creatures you control this turn."
//
// The hexproof is Dawn's Truce's: the player keyword for you, and a
// keyword grant over the permanents you control as Surge resolves (it
// modifies characteristics, so the set is fixed then, CR 611.2c).
//
// The shield is #2045's recipient set beside ADR 0108 §7's property
// recheck, in one record: the source is black or red (or both) as it
// would deal the damage (CR 615.9), and the recipient is a creature you
// control at that moment. A prevention effect doesn't modify
// characteristics, so a creature that enters later is protected; damage
// to you is not prevented.
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b9a62c6b-5848-44b1-932b-8d5de2560976",
		Name:         "Surge of Salvation",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			if err := (effects.GainPlayerKeyword{
				Player:   ctx.Controller(),
				Keyword:  effects.KeywordHexproof,
				Label:    "Surge of Salvation — hexproof",
				Duration: effects.DurationUntilEndOfTurn(ctx),
			}).Apply(ctx); err != nil {
				return err
			}
			if err := (effects.GrantKeywordUntilEOT{
				Match:    effects.YouControl(),
				Keywords: []string{"hexproof"},
				Label:    "Surge of Salvation — permanents you control gain hexproof",
			}).Apply(ctx); err != nil {
				return err
			}
			return effects.PreventDamageFromSource{
				Protect: effects.ShieldCreaturesYouControl,
				Queries: []game.PermanentQuery{effects.QueryColors("B", "R")},
			}.Apply(ctx)
		},
	})
}
