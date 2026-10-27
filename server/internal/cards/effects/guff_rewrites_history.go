package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Guff Rewrites History — Instant {2}{R}:
//
//	"For each player, choose target nonenchantment, nonland permanent
//	 that player controls. Those permanents' owners shuffle them into
//	 their libraries. Each player who controlled one of those permanents
//	 exiles cards from the top of their library until they exile a
//	 nonland card, then puts the rest on the bottom of their library in
//	 a random order. Each player may cast the nonland card they exiled
//	 without paying its mana cost."
//
// THE TARGETS. One required clause per player, each bound to its
// player: the caster's own ("you control"), then one per opponent seat
// (ForNthOpponent, Blatant Thievery's shape, #4128). A seat with no
// live opponent is not required; a player at the table who controls no
// nonenchantment, nonland permanent leaves their clause unfillable, so
// the spell can't be cast then, the reading Blatant Thievery takes of
// CR 601.2c. Each pick is re-checked against its own player on
// resolution (CR 608.2b), so one that changed hands is skipped.
//
// THE SHUFFLE. Every still-legal target goes to its owner's library at
// once (TuckCardsToLibraryKeyedForEffect, one simultaneous move; a
// commander may take CR 903.9's offer instead), then each owner of one
// shuffles. The players who controlled one are read before the move.
//
// THE WALK. Each of those players, in APNAP order, exiles until a
// nonland card and puts the rest on the bottom in a random order, then
// is asked whether to cast that card free
// (ExileUntilCastFreeKeepingForEffect, cascade's body with one
// difference). The cast is made inside the spell's resolution, as
// cascade's and Tibalt's Trickery's are: a resolution_cast prompt for
// each player over their card (CR 608.2g, #4250).
//
// A card the player declines stays in exile as printed; only the rest of
// the pile goes to the bottom (#4205).
// Guffster's Party deck request (#4184).
func init() {
	Register(Spec{
		OracleID:     "7ff7af5f-e4e3-40af-8882-756564bfa479",
		Name:         "Guff Rewrites History",
		Completeness: CompletenessFull,
		Targets:      guffRewritesHistoryClauses(),
		OnResolve:    guffRewritesHistoryResolve,
	})
}

// guffRewritesHistoryClauses is "for each player, choose target
// nonenchantment, nonland permanent that player controls".
func guffRewritesHistoryClauses() *game.TargetSpec {
	mine := TargetPermanent("target nonenchantment, nonland permanent you control",
		Not(Enchantment()), Nonland(), YouControl())
	var rest []*game.TargetSpec
	for n := 1; n <= game.MaxPlayers-1; n++ {
		rest = append(rest, ForNthOpponent(n, TargetPermanent(
			"target nonenchantment, nonland permanent controlled by the "+nthOpponentWords[n]+" opponent after you in turn order",
			Not(Enchantment()), Nonland())))
	}
	return Clauses(mine, rest...)
}

func guffRewritesHistoryResolve(item *game.StackItem, ctx *Context) error {
	var ids []uuid.UUID
	controlled := map[uuid.UUID]bool{}
	owners := map[uuid.UUID]bool{}
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard || !ctx.Game.Battlefield.Contains(t.ID) {
			continue
		}
		c, ok := ctx.Game.LookupCardForEffect(t.ID)
		if !ok {
			continue
		}
		ids = append(ids, t.ID)
		controlled[c.Controller] = true
		owners[c.Owner] = true
	}
	if len(ids) == 0 {
		return nil
	}
	// Keyed (#4253 slice 10a): Carry is the owners, then the controllers;
	// N is how many owners there are.
	var carry []uuid.UUID
	for p := range owners {
		carry = append(carry, p)
	}
	n := len(carry)
	for p := range controlled {
		carry = append(carry, p)
	}
	tail := game.KeyedTail{Ref: guffShuffleThenExile, Player: item.Controller, Source: item.SourceCardID, Carry: carry, N: n}
	return ctx.Game.TuckCardsToLibraryKeyedForEffect(ids, game.TuckOptions{}, tail)
}

// guffShuffleThenExile is the tuck's "then each owner shuffles, then each
// controller exiles cards until a nonland card and may cast it", in APNAP
// order (#4253 slice 10a).
var guffShuffleThenExile = game.RegisterTail("then/guff-rewrites-history-shuffle-exile", func(g *game.Game, o game.TailOutcome) error {
	n := min(max(o.Tail.N, 0), len(o.Tail.Carry))
	owners, controlled := map[uuid.UUID]bool{}, map[uuid.UUID]bool{}
	for _, p := range o.Tail.Carry[:n] {
		owners[p] = true
	}
	for _, p := range o.Tail.Carry[n:] {
		controlled[p] = true
	}
	order := apnapPlayers(g)
	for _, p := range order {
		if owners[p] {
			if err := g.ShuffleLibraryForEffect(p); err != nil {
				return err
			}
		}
	}
	for _, p := range order {
		if !controlled[p] {
			continue
		}
		if err := g.ExileUntilCastFreeKeepingForEffect(p, o.Tail.Source, func(c game.Card) bool { return !c.IsLand() }); err != nil {
			return err
		}
	}
	return nil
})
