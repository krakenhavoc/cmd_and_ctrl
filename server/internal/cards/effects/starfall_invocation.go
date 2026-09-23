package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Starfall Invocation — Sorcery {3}{W}{W} (EDHREC rank 2210):
//
//	"Gift a card (You may promise an opponent a gift as you cast this
//	 spell. If you do, they draw a card before its other effects.)
//	 Destroy all creatures. If the gift was promised, return a
//	 creature card put into your graveyard this way to the battlefield
//	 under your control."
//
// Wrath of God that can keep its best creature for the price of a
// card to an opponent. The wipe is DestroyAllMatching (S23): every
// creature dies as one event, so the aristocrats payoffs see all of
// them.
//
// Sandbox simplification, declared: the GIFT is not offered. Gift
// itself exists since ADR 0089 (#1267) — `Gift: GiftACard()` would
// offer the promise and draw the opponent a card — but the promised
// branch needs a pick among the creature cards THIS wipe put into
// your graveyard, and a card that offered the promise without the
// return would be a trap: the opponent draws and you get nothing. So
// the spell stays its base mode — destroy all creatures — which is the
// printed card with one option removed, weaker and never stronger.
// Declaring the gift is one line the day the pick lands.
//
// The catalog-wide one this used to carry alongside it (#446 — the
// mass destroy path not consulting indestructible) was the engine
// gap it said it was, and S30 (#470) landed the fix.
func init() {
	Register(Spec{
		OracleID:     "7024532b-f99b-43a7-b0ed-5b3e7ec7592b",
		Name:         "Starfall Invocation",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"The gift can't be promised, so the spell only destroys all creatures — it never returns one of yours to the battlefield.",
		},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return DestroyAllMatching{Match: Creature()}.Apply(ctx)
		},
	})
}
