package cardsso

import (
	"strconv"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Soul Strings — Sorcery {X}{B}:
//
//	"Return two target creature cards from your graveyard to your hand
//	 unless any player pays {X}."
//
// ADR 0150 (#1903): {X} is the X the spell was cast with, offered to
// every player in APNAP order from the active player (AnyPlayerPays).
// With X of 0 the first player who says "pay" stops it for nothing,
// which is the printed card. A target that left the graveyard while
// the offer was open is skipped (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fac310c7-c1e5-42f9-8140-17472be21d0d",
		Name:         "Soul Strings",
		Completeness: effects.CompletenessFull,
		// X is what each player is asked to pay; at X=0 anyone stops it
		// for nothing.
		XMatters: true,
		Targets:  effects.TargetCardInGraveyard("two target creature cards from your graveyard", effects.YouOwn(), effects.Creature()).WithCount(2, 2),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			cost := "{" + strconv.Itoa(ctx.X()) + "}"
			return effects.AnyPlayerPays{
				Cost:     cost,
				Question: "Soul Strings — pay " + cost + " to stop the return?",
				Then:     soulStringsReturns,
			}.Apply(ctx)
		},
	})
}

// soulStringsReturns is "if no one does", keyed (#4264): the spell's targets ride the asking item.
var soulStringsReturns = effects.AnyPlayerConsequences("any-player/soul-strings-return",
	nil,
	func(ctx *effects.Context, _ game.AnyPlayerAnswer) error {
		return effects.ReturnLegalGraveyardTargetsToHand(ctx)
	},
	nil)
