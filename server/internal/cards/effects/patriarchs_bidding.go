package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Patriarch's Bidding — Sorcery {3}{B}{B}:
//
//	"Each player chooses a creature type. Each player returns all
//	 creature cards of a type chosen this way from their graveyard to
//	 the battlefield."
//
// "Each player chooses" is a chain of resolution-time prompts (#2382),
// one per living seat in turn order from the caster, each addressed to
// that seat. The chosen types ride the continuation BY VALUE. When the
// last seat has answered, every player returns every creature card of
// any chosen type from their own graveyard, as one simultaneous entry
// under its owner's control. A seat that leaves with its prompt open
// chooses nothing and the chain carries on.
//
// The prompts are asked one at a time, caster first, so a later seat
// sees the earlier answers — which is the rules' own order for an
// "each player chooses" (CR 101.4: APNAP, with earlier choices known).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "25fc3bc2-d852-4f52-9adb-c7e35c06f3af",
		Name:         "Patriarch's Bidding",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return patriarchsBiddingAsk(ctx.Game, item, seatsFromController(ctx.Game, item.Controller), nil)
		},
	})
}

// patriarchsBiddingAsk asks seats[0], then recurses on the rest with
// the answer appended. A seat that cannot be asked (queue refused) is
// skipped without an answer.
func patriarchsBiddingAsk(g *game.Game, item *game.StackItem, seats []uuid.UUID, chosen []string) error {
	if len(seats) == 0 {
		return patriarchsBiddingReturn(NewContext(g, item), chosen)
	}
	rest := seats[1:]
	asked := ChooseCreatureTypeThen(g, seats[0], item.SourceCardID, "Patriarch's Bidding — choose a creature type",
		func(g *game.Game, t string) error {
			next := chosen
			if t != "" {
				next = append(append([]string(nil), chosen...), t)
			}
			return patriarchsBiddingAsk(g, item, rest, next)
		})
	if !asked {
		return patriarchsBiddingAsk(g, item, rest, chosen)
	}
	return nil
}

func patriarchsBiddingReturn(ctx *Context, chosen []string) error {
	if len(chosen) == 0 {
		return nil
	}
	ids := allGraveyardsCardIDs(ctx, func(c game.Card) bool {
		if !c.IsCreature() {
			return false
		}
		for _, t := range chosen {
			if c.HasSubtype(t) {
				return true
			}
		}
		return false
	})
	return ReturnFromGraveyardTogether{Targets: ids}.Apply(ctx)
}
