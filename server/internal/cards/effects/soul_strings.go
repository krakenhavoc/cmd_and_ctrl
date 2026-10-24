package effects

import (
	"strconv"

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
	Register(Spec{
		OracleID:     "fac310c7-c1e5-42f9-8140-17472be21d0d",
		Name:         "Soul Strings",
		Completeness: CompletenessFull,
		// X is what each player is asked to pay; at X=0 anyone stops it
		// for nothing.
		XMatters: true,
		Targets:  TargetCardInGraveyard("two target creature cards from your graveyard", YouOwn(), Creature()).WithCount(2, 2),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			cost := "{" + strconv.Itoa(ctx.X()) + "}"
			return AnyPlayerPays{
				Cost:     cost,
				Question: "Soul Strings — pay " + cost + " to stop the return?",
				Then:     soulStringsReturns,
			}.Apply(ctx)
		},
	})
}

// soulStringsReturns is "if no one does", keyed (#4264): the spell's targets ride the asking item.
var soulStringsReturns = AnyPlayerConsequences("any-player/soul-strings-return",
	nil,
	func(ctx *Context, _ game.AnyPlayerAnswer) error { return returnLegalGraveyardTargetsToHand(ctx) },
	nil)
