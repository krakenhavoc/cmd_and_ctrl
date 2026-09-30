package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Smuggler's Surprise — Instant {G} (#1112):
//
//	"Spree (Choose one or more additional costs.)
//	 + {2} — Mill four cards. You may put up to two creature and/or
//	   land cards from among the milled cards into your hand.
//	 + {4}{G} — You may put up to two creature cards from your hand
//	   onto the battlefield.
//	 + {1} — Creatures you control with power 4 or greater gain
//	   hexproof and indestructible until end of turn."
//
// Spree (CR 702.172a, ADR 0065's 2026-09-23 amendment), with the
// bullets declared by SpreeMode and resolved by ONE OnResolve in
// printed order (CR 608.2c) rather than by per-bullet bodies. The
// bullets see each other, and the order is the card:
//
//   - a creature card the first bullet puts into your hand is one the
//     second bullet may put onto the battlefield, so the second bullet
//     asks only once the first bullet's pick is in;
//   - a creature the second bullet puts onto the battlefield with power
//     4 or greater is one the third bullet protects, so the third
//     bullet's set is fixed (CR 611.2c) only once the second bullet's
//     creatures have entered.
//
// Both picks are asynchronous prompts, so each bullet hands the next
// one over as a continuation — the next line would run before the
// player had answered. Every continuation captures only the resolving
// stack item and rebuilds its Context from the game it is handed, so
// an undo across a prompt resolves against the restored game.
//
// "From among the milled cards" is MillToZone.Then's list — the cards
// that actually reached the graveyard — re-checked against the
// graveyard when the prompt is built (mill_then_take.go's reasoning).
// The two creature cards from hand enter as ONE simultaneous event
// (PutOntoBattlefieldTogetherThenForEffect), so each sees the other
// arrive (CR 603.6a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "605f719d-01f8-406c-9d4c-3f4992c6a69f",
		Name:         "Smuggler's Surprise",
		Completeness: CompletenessFull,
		Modes: Spree(
			SpreeMode("Mill four cards. You may put up to two creature and/or land cards from among the milled cards into your hand.", "{2}"),
			SpreeMode("You may put up to two creature cards from your hand onto the battlefield.", "{4}{G}"),
			SpreeMode("Creatures you control with power 4 or greater gain hexproof and indestructible until end of turn.", "{1}"),
		),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return smugglersSurpriseMill(ctx.Game, item)
		},
	})
}

// smugglersSurpriseMill is the first bullet, then the rest.
func smugglersSurpriseMill(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if !ctx.HasMode(0) {
		return smugglersSurprisePut(g, item)
	}
	return MillToZone{
		Player: item.Controller,
		N:      4,
		Then: func(ctx *Context, milled []uuid.UUID) error {
			return smugglersSurpriseTakeMilled(ctx.Game, ctx.Item, milled)
		},
	}.Apply(ctx)
}

// smugglersSurpriseTakeMilled offers up to two creature and/or land
// cards from among the milled ones, and moves on to the second bullet
// once the answer is in.
func smugglersSurpriseTakeMilled(g *game.Game, item *game.StackItem, milled []uuid.UUID) error {
	player := item.Controller
	var candidates []uuid.UUID
	for _, id := range milled {
		c, ok := g.LookupCardForEffect(id)
		if !ok || !(c.IsCreature() || c.IsLand()) {
			continue
		}
		if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneGraveyard {
			continue
		}
		candidates = append(candidates, id)
	}
	if len(candidates) == 0 {
		return smugglersSurprisePut(g, item)
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:    player,
		FromPlayer: player,
		Source:     item.SourceCardID,
		Question:   "Smuggler's Surprise — put up to two creature and/or land cards from among the milled cards into your hand",
		Cards:      candidates,
		Min:        0,
		Max:        2,
		Zone:       game.ZoneGraveyard,
		Then: func(g *game.Game, picked []uuid.UUID) error {
			if err := ReturnPickedToHand(item)(g, picked); err != nil {
				return err
			}
			return smugglersSurprisePut(g, item)
		},
	})
	return nil
}

// smugglersSurprisePut is the second bullet: up to two creature cards
// from hand, entering together, then the third bullet.
func smugglersSurprisePut(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if !ctx.HasMode(1) {
		return smugglersSurpriseProtect(g, item)
	}
	player := item.Controller
	candidates := handCardsMatching(g, player, Creature())
	if len(candidates) == 0 {
		return smugglersSurpriseProtect(g, item)
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  player,
		Source:   item.SourceCardID,
		Question: "Smuggler's Surprise — put up to two creature cards from your hand onto the battlefield",
		Cards:    candidates,
		Min:      0,
		Max:      2,
		Zone:     game.ZoneHand,
		Then: func(g *game.Game, picked []uuid.UUID) error {
			if len(picked) == 0 {
				return smugglersSurpriseProtect(g, item)
			}
			entries := make([]game.BatchEntry, 0, len(picked))
			for _, id := range picked {
				entries = append(entries, game.BatchEntry{CardID: id, From: game.ZoneHand})
			}
			return g.PutOntoBattlefieldTogetherThenForEffect(entries, game.ZoneEntryOptions{Controller: player},
				func(g *game.Game, _ []uuid.UUID) error {
					return smugglersSurpriseProtect(g, item)
				})
		},
	})
	return nil
}

// smugglersSurpriseProtect is the third bullet. The affected set is
// the creatures you control with power 4 or greater NOW (CR 611.2c),
// after the first two bullets have finished.
func smugglersSurpriseProtect(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if !ctx.HasMode(2) {
		return nil
	}
	return GrantKeywordUntilEOT{
		Match:    And(Creature(), YouControl(), PowerGE(4)),
		Keywords: []string{"hexproof", "indestructible"},
		Label:    "Smuggler's Surprise — hexproof and indestructible until end of turn",
	}.Apply(ctx)
}
