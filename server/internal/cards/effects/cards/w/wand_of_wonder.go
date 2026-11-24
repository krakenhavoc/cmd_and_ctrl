package cardsw

import (
	"strconv"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Wand of Wonder — Artifact {3}{R} (Adventures in the Forgotten Realms):
//
//	"{4}, {T}: Roll a d20. Each opponent exiles cards from the top of their
//	 library until they exile an instant or sorcery card, then shuffles the
//	 rest into their library. You may cast up to X instant and/or sorcery
//	 spells from among cards exiled this way without paying their mana costs.
//	 1—9 | X is one.
//	 10—19 | X is two.
//	 20 | X is three."
//
// The die is rolled first (CR 706.2) and the results table (CR 706.3a) is read
// from the natural result, so a replaced or modified roll is the result the table
// sees. Each opponent then exiles in turn order from the active player (Plargg
// and Nassari's walk): a run of single exiles that stops after the first
// instant or sorcery card (UntilInstantOrSorcery), each library one waiting on
// the one before. The exiled cards that are not that card, or all of them when
// the library holds no instant or sorcery, are put into their owner's library and
// it is shuffled (CR 701.24); a library with a commander in the pile offers its
// owner the command zone as usual. The cards that stay exiled are the hits, one
// per opponent.
//
// The casts are one permission over the hits good for X casts (CastPermission.
// CastsLeft, Plargg and Nassari's two): made during the resolution (CR 608.2g),
// each offered in turn through the resolution_cast prompt until X are made or the
// player declines. A spell with {X} in its cost is cast with X = 0 (CR 107.3b) and
// a mandatory additional cost is still paid. A hit that is not cast stays exiled.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8f1123dd-19b7-402b-8ecd-ab62494c9fd5",
		Name:         "Wand of Wonder",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{4}, {T}: Roll a d20. Each opponent exiles cards from the top of their library until they exile an instant or sorcery card, then shuffles the rest into their library. You may cast up to X instant and/or sorcery spells from among cards exiled this way without paying their mana costs.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{4}"), effects.TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.RollDice{Sides: 20, N: 1, Then: wandOfWonderTable}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}

const wandOfWonderName = "Wand of Wonder"

// The two clauses refer to the walk and the walk to them, so they are registered
// in init rather than by package variable initialisers.
var wandOfWonderExiled, wandOfWonderTucked game.TailRef

func init() {
	wandOfWonderExiled = effects.CardTail("wand-of-wonder/after-exile", wandOfWonderAfterExile)
	wandOfWonderTucked = effects.CardTail("wand-of-wonder/after-tuck", wandOfWonderAfterTuck)
}

// wandOfWonderTable is the results table: X from the roll, then the first
// opponent's exile. An on-disk key.
var wandOfWonderTable = effects.DiceTail("dice/wand-of-wonder-table", func(ctx *effects.Context, rolls []int) error {
	x := 1
	switch {
	case rolls[0] >= 20:
		x = 3
	case rolls[0] >= 10:
		x = 2
	}
	var opps []uuid.UUID
	for _, id := range effects.ApnapLivingSeats(ctx.Game) {
		if id != ctx.Controller() {
			opps = append(opps, id)
		}
	}
	return wandOfWonderWalk(ctx, x, opps, nil)
})

// wandOfWonderWalk exiles for the first of `opps` and then walks on to the rest;
// `hits` are the instants and sorceries exiled so far. With no opponent left the
// casts are offered.
func wandOfWonderWalk(ctx *effects.Context, x int, opps, hits []uuid.UUID) error {
	if len(opps) == 0 {
		return wandOfWonderOffer(ctx.Game, ctx.Controller(), ctx.Source(), x, hits)
	}
	rest := opps[1:]
	then := effects.WithN(effects.WithCarry(ctx.ItemTail(wandOfWonderExiled), append(append([]uuid.UUID(nil), rest...), hits...)...), len(rest))
	then.Player = opps[0]
	then.Words = []string{strconv.Itoa(x)}
	return effects.MillToZone{Player: opps[0], To: game.ZoneExile, Until: effects.UntilInstantOrSorcery, ThenTail: then}.Apply(ctx)
}

// wandOfWonderSplit reads a tail's carried state: the opponents still to walk,
// the hits so far and X.
func wandOfWonderSplit(o game.TailOutcome) (opps, hits []uuid.UUID, x int) {
	n := min(max(o.Tail.N, 0), len(o.Tail.Carry))
	x = 1
	if len(o.Tail.Words) > 0 {
		if v, err := strconv.Atoi(o.Tail.Words[0]); err == nil {
			x = v
		}
	}
	return o.Tail.Carry[:n], o.Tail.Carry[n:], x
}

// wandOfWonderExiled is the clause after one opponent's exile: the last card
// exiled is the hit when it is an instant or sorcery, the rest go back into
// that player's library, which is shuffled, and the next opponent is walked. An
// on-disk key.
func wandOfWonderAfterExile(ctx *effects.Context, o game.TailOutcome) error {
	opps, hits, x := wandOfWonderSplit(o)
	var rest []uuid.UUID
	for i, id := range o.IDs {
		c, ok := ctx.Game.LookupCardForEffect(id)
		if i == len(o.IDs)-1 && ok && !c.IsToken() && (c.IsInstant() || c.IsSorcery()) {
			hits = append(append([]uuid.UUID(nil), hits...), id)
			continue
		}
		rest = append(rest, id)
	}
	if len(rest) == 0 {
		return wandOfWonderWalk(ctx, x, opps, hits)
	}
	then := effects.WithN(effects.WithCarry(ctx.ItemTail(wandOfWonderTucked), append(append([]uuid.UUID(nil), opps...), hits...)...), len(opps))
	then.Player = o.Tail.Player
	then.Words = o.Tail.Words
	return ctx.Game.TuckCardsToLibraryKeyedForEffect(rest, game.TuckOptions{}, then)
}

// wandOfWonderTucked is the clause once the rest is back in the library: it is
// shuffled (CR 701.24), then the next opponent is walked. An on-disk key.
func wandOfWonderAfterTuck(ctx *effects.Context, o game.TailOutcome) error {
	if err := ctx.Game.ShuffleLibraryForEffect(o.Tail.Player); err != nil {
		return err
	}
	opps, hits, x := wandOfWonderSplit(o)
	return wandOfWonderWalk(ctx, x, opps, hits)
}

// wandOfWonderOffer grants the one permission over the hits still in exile,
// good for `x` casts without paying mana costs, and offers it during the
// resolution.
func wandOfWonderOffer(g *game.Game, controller, source uuid.UUID, x int, hits []uuid.UUID) error {
	var cards []game.Card
	var ids []uuid.UUID
	for _, id := range hits {
		z := g.FindCardZoneForEffect(id)
		c, ok := g.LookupCardForEffect(id)
		if !ok || z == nil || z.Kind != game.ZoneExile {
			continue
		}
		cards = append(cards, c)
		ids = append(ids, id)
	}
	if len(cards) == 0 {
		return nil
	}
	g.GrantCastPermissionToCardsForEffect(game.CastPermission{
		Player:      controller,
		Zone:        game.ZoneExile,
		Duration:    g.UntilEndOfTurnDuration(),
		Cost:        "{0}",
		Timing:      game.TimingFlash,
		CastOnly:    true,
		CastsLeft:   x,
		AnyNumber:   true,
		LapseOnPass: game.LapseStaysInExile,
		Source:      source,
		SourceName:  wandOfWonderName,
		Label:       wandOfWonderName + " — cast up to " + strconv.Itoa(x) + " of the exiled instants and sorceries without paying their mana costs",
	}, cards)
	effects.OfferResolutionCastOver(g, controller, source, wandOfWonderName, game.MayCastKeywordFreeCast, effects.DeclineLeaveInExile, ids...)
	return nil
}
