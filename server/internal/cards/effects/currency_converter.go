package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Currency Converter — Artifact {1} (batch 25, #387):
//
//	"Whenever you discard a card, you may exile that card from your
//	 graveyard.
//	 {2}, {T}: Draw a card, then discard a card.
//	 {T}: Put a card exiled with this artifact into its owner's
//	 graveyard. If it's a land card, create a Treasure token. If it's
//	 a nonland card, create a 2/2 black Rogue creature token."
//
// A looter that turns its own discards into a slow trickle of
// Treasures and Rogues instead of graveyard value.
//
// # This card needed nothing new either
//
// Declared skipped on batch 25 for "an 'exiled with this' record plus
// an exile-to-graveyard move" — the same seam Valakut Exploration sat
// on, and the same answer: b27ExiledWith (already shipped for
// Duplicant, Angel of Serenity, Chrome Mox, Bag of Holding,
// Ossification) plus ordinary primitives. See the doc comment on
// valakut_exploration.go for the full accounting; #1218's actual new
// primitive is The Ozolith's counters-at-departure LKI, which this
// card does not touch.
//
// The trigger is Bag of Holding's exile trigger with ONE difference:
// "you may" here, unconditional there — Optional wraps the same
// AppliesTo/Build shape (bag_of_holding.go), and both read ev.CardID
// captured at Build time since a discard names exactly one card. The
// resolution body itself, exileFromGraveyardIfStillThere, is shared
// verbatim (helpers.go) rather than copied — the two were an exact
// duplicate before that extraction.
//
// The third ability is what Bag of Holding and its siblings did not
// need: a CHOICE among the exiled-with pile rather than "all of them"
// or "the newest one". game.ChooseCardsPrompt already generalizes
// "choose N from a computed pool" (the discard prompt, the search
// prompt, the untap-choice prompt) — b27ExiledWith's own result
// slots in as the pool with no new plumbing, Min 0 so activating with
// nothing exiled just does nothing (CR 602.2b: the ability is a
// legal activation either way, it simply has no legal choice to
// offer).
//
// No simplification.
const currencyConverterOracle = "981298e6-ddee-49c0-9377-f47f019b4138"

// currencyConverterExileLabel is the stack label the discard trigger
// carries. b27ExiledWith keys the "exiled with this artifact" record
// on it, so the trigger and the third ability must agree.
const currencyConverterExileLabel = "Currency Converter — exile that card from your graveyard"

func init() {
	Register(Spec{
		OracleID:     currencyConverterOracle,
		Name:         "Currency Converter",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Optional(game.TriggeredAbility{
				Watches: []game.EventKind{game.EventDiscardCard},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return discardedByYou(ev, source)
				},
				Build: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
					discarded := ev.CardID
					return game.NewTriggeredItem(source, currencyConverterExileLabel, func(g *game.Game, item *game.StackItem) error {
						return exileFromGraveyardIfStillThere(g, item, discarded)
					})
				},
			}, "Currency Converter — exile that card from your graveyard?"),
		},
		Activated: []ActivatedAbility{
			{
				Label: "{2}, {T}: Draw a card, then discard a card.",
				Cost:  Plus(ManaCost("{2}"), TapCost()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return b16DrawThenDiscard(g, item, 1, 1)
				},
			},
			{
				Label:  "{T}: Put a card exiled with this artifact into its owner's graveyard. If it's a land card, create a Treasure token. If it's a nonland card, create a 2/2 black Rogue creature token.",
				Cost:   TapCost(),
				Effect: currencyConverterPutAwayAnExiledCard,
			},
		},
	})
}

// currencyConverterPutAwayAnExiledCard is the third ability's body:
// offer a choice among the cards exiled with this artifact (empty
// pool is a legal, silent no-op), then put the chosen one into its
// owner's graveyard and make the matching token off its PRINTED type
// line — read where it sits in exile, since a card off the
// battlefield has no layered characteristic.
func currencyConverterPutAwayAnExiledCard(g *game.Game, item *game.StackItem) error {
	ids := b27ExiledWith(g, item.SourceCardID, currencyConverterExileLabel)
	if len(ids) == 0 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  item.Controller,
		Source:   item.SourceCardID,
		Question: "Currency Converter — put a card exiled with this artifact into its owner's graveyard",
		Cards:    ids,
		Zone:     game.ZoneExile,
		Min:      0,
		Max:      1,
		Then: func(g *game.Game, picked []uuid.UUID) error {
			if len(picked) == 0 {
				return nil
			}
			chosen, ok := g.LookupCardForEffect(picked[0])
			if !ok {
				return nil
			}
			isLand := chosen.IsLand()
			return g.PutIntoGraveyardThenForEffect(picked[0], func(g *game.Game, landed bool) error {
				if !landed {
					return nil
				}
				ctx := NewContext(g, item)
				if isLand {
					return CreateToken{Controller: item.Controller, Template: TreasureToken(), N: 1}.Apply(ctx)
				}
				return CreateToken{Controller: item.Controller, Template: TokenCard("2/2 black Rogue"), N: 1}.Apply(ctx)
			})
		},
	})
	return nil
}
