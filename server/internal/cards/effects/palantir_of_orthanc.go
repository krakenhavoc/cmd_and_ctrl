package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Palantír of Orthanc — Legendary Artifact {3} (Lord of the Rings):
//
//	"At the beginning of your end step, put an influence counter on
//	 Palantír of Orthanc and scry 2. Then target opponent may have you
//	 draw a card. If that player doesn't, you mill X cards, where X is
//	 the number of influence counters on Palantír of Orthanc, and that
//	 player loses life equal to the total mana value of those cards."
//
// A repeating, growing Combustible Gearhulk: the opponent's yes/no
// decision is MayChoice with Player addressing (#796), and the "if
// the player doesn't" branch mills and burns off the cards that
// actually moved (CR 701.17b — a short library mills what it has),
// exactly as Combustible Gearhulk's does. What is new here is that
// the mill count is not a printed constant but a live read of the
// influence counters, and there are three sequential decisions in one
// trigger:
//
//  1. The counter goes on FIRST (so X for this activation already
//     includes it), then scry 2 queues its own prompt.
//  2. The opponent's yes/no can only be asked once the scry is
//     settled — Scry's Then is exactly the primitive for "queue the
//     next decision once this one is answered", the same ordering
//     rule Preordain's "scry, then draw" follows.
//  3. "That player" in both the draw and the life-loss clauses is the
//     SAME opponent the trigger targeted; CR 608.2b means an opponent
//     who leaves the game (or is no longer legal) in response gets no
//     question at all rather than an automatic "doesn't".
//
// "Influence counter" is a free-form counter kind (Card.Counters is
// string-keyed); no engine change needed to add a new named counter.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7798a4d9-1c1d-48e6-b9a2-ecd6aec1efa7",
		Name:         "Palantír of Orthanc",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				AtYourEndStep("Palantír of Orthanc — an influence counter, scry 2, then an opponent's choice",
					palantirTrigger),
				TargetPlayer("target opponent", Opponent()),
			),
		},
	})
}

func palantirTrigger(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if err := (AddCounter{Target: item.SourceCardID, Kind: "influence", N: 1}).Apply(ctx); err != nil {
		return err
	}
	return Scry{Player: item.Controller, N: 2, Then: func(g *game.Game) error {
		return palantirAskOpponent(g, item)
	}}.Apply(ctx)
}

func palantirAskOpponent(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetPlayer {
			continue
		}
		opponent := t.ID
		return MayChoice{
			Player:   opponent,
			Question: "Palantír of Orthanc — let its controller draw a card?",
			YesLabel: "They draw a card",
			NoLabel:  "They mill instead and I lose life",
			OnYes: func(ctx *Context) error {
				return DrawCards{Player: item.Controller, N: 1}.Apply(ctx)
			},
			OnNo: func(ctx *Context) error {
				return palantirMillAndBurn(ctx, item, opponent)
			},
		}.Apply(ctx)
	}
	// CR 608.2b: the target opponent is gone, so nothing happens at
	// all — not the "doesn't" branch, which would burn a player who
	// was never asked.
	return nil
}

// palantirMillAndBurn is the "if that player doesn't" branch: mill X
// cards, where X is the LIVE influence-counter count, then that
// opponent loses life equal to their total mana value.
func palantirMillAndBurn(ctx *Context, item *game.StackItem, opponent uuid.UUID) error {
	c, ok := ctx.Game.LookupCardForEffect(item.SourceCardID)
	if !ok {
		// Palantír left the battlefield between the trigger firing and
		// the opponent's answer — nothing to mill by, so the ability
		// does as much as it can (CR 608.2c), which is nothing.
		return nil
	}
	x := c.Counters["influence"]
	if x <= 0 {
		return nil
	}
	return ctx.Game.MillToZoneThenForEffect(item.Controller, x, game.ZoneGraveyard, nil,
		func(g *game.Game, milled []uuid.UUID) error {
			total := 0
			for _, id := range milled {
				if mc, ok := g.LookupCardForEffect(id); ok {
					total += mc.ManaValue()
				}
			}
			if total <= 0 {
				return nil
			}
			return g.ChangePlayerLifeForEffect(item.SourceCardID, opponent, -total)
		})
}
