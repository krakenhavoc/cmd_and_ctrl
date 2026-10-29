package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

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
// The gift is `Gift: GiftACard()` (ADR 0089): the opponent draws before
// the spell's other effects. The return is the sweep's own `Then`
// clause, which hears the creatures actually destroyed; the candidates
// are the creature CARDS among them that now sit in the caster's
// graveyard, so a token (it ceases to exist), a commander that went to
// the command zone and a creature that was exiled instead are not
// offered. One candidate returns outright, several are offered as an
// exactly-one pick, and the creature comes back under the caster's
// control.
//
// The catalog-wide one this used to carry alongside it (#446 — the
// mass destroy path not consulting indestructible) was the engine
// gap it said it was, and S30 (#470) landed the fix.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7024532b-f99b-43a7-b0ed-5b3e7ec7592b",
		Name:         "Starfall Invocation",
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy}},
		Completeness: CompletenessFull,
		Gift:         GiftACard(),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			promised := 0
			if ctx.GiftPromised() {
				promised = 1
			}
			return DestroyAllMatching{Match: Creature(), ThenTail: withN(ctx.ItemTail(starfallReturn), promised)}.Apply(ctx)
		},
	})
}

// starfallReturn is the sweep's "if the gift was promised, return …",
// keyed (#4253 slice 10a): N is 1 when the gift was promised.
var starfallReturn = MassTailOf("mass/starfall-invocation-return", func(ctx *Context, swept []game.Card, _ int, t game.KeyedTail) error {
	if t.N == 0 {
		return nil
	}
	return starfallReturnOne(ctx, ctx.Item, swept)
})

// starfallReturnOne is "return a creature card put into your graveyard
// this way to the battlefield under your control": the destroyed
// creature cards the caster owns that really are in their graveyard.
func starfallReturnOne(ctx *Context, item *game.StackItem, swept []game.Card) error {
	owner := item.Controller
	was := map[uuid.UUID]bool{}
	for _, c := range swept {
		if c.Owner == owner && !IsToken(c) {
			was[c.InstanceID] = true
		}
	}
	ids := graveyardCardIDs(ctx, owner, func(c game.Card) bool { return was[c.InstanceID] && c.IsCreature() })
	return b17ReanimateOneMilled(ctx.Game, item, owner,
		"Starfall Invocation — return a creature card put into your graveyard this way", ids)
}
