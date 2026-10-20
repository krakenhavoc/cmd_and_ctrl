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
// once (TuckCardsToLibraryThenForEffect, one simultaneous move; a
// commander may take CR 903.9's offer instead), then each owner of one
// shuffles. The players who controlled one are read before the move.
//
// THE WALK. Each of those players, in APNAP order, exiles until a
// nonland card and puts the rest on the bottom in a random order, then
// is asked whether to cast that card free (ExileUntilCastFreeForEffect,
// cascade's body). The cast uses the shared stand-in cascade and
// Tibalt's Trickery use: a free grant taken right after the spell
// resolves rather than a cast inside its resolution.
//
// One difference from the printed card, and the reason for the caveat:
// a card the player declines (or doesn't cast before passing) goes to
// the bottom of their library with the rest; printed, it stays in
// exile. Guffster's Party deck request (#4184).
func init() {
	Register(Spec{
		OracleID:     "7ff7af5f-e4e3-40af-8882-756564bfa479",
		Name:         "Guff Rewrites History",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"A card a player chooses not to cast goes to the bottom of their library instead of staying in exile."},
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
	source := item.SourceCardID
	return ctx.Game.TuckCardsToLibraryThenForEffect(ids, game.TuckOptions{}, func(g *game.Game, _ []uuid.UUID) error {
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
			if err := g.ExileUntilCastFreeForEffect(p, source, func(c game.Card) bool { return !c.IsLand() }); err != nil {
				return err
			}
		}
		return nil
	})
}
