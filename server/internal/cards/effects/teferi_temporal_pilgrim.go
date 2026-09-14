package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Teferi, Temporal Pilgrim — Legendary Planeswalker — Teferi for
// {3}{U}{U}, starting loyalty 4 (EDHREC rank 3342):
//
//	"Whenever you draw a card, put a loyalty counter on Teferi.
//	 0: Draw a card.
//	 −2: Create a 2/2 blue Spirit creature token with vigilance and
//	     'Whenever you draw a card, put a +1/+1 counter on this
//	     token.'
//	 −12: Target opponent chooses a permanent they control and
//	      returns it to its owner's hand. Then they shuffle each
//	      nonland permanent they control into its owner's library."
//
// The static-looking first line is a triggered ability and it is
// complete: every draw its controller makes — their draw step, their
// Rhystic Study, or Teferi's own [0] — puts a loyalty counter on.
// The counter is placed by an EFFECT rather than paid as a cost, so
// Doubling Season and Vorinclex do apply to it, which is the
// opposite of what happens to the [0] itself (ADR 0032 §8).
//
// The [0] is the reason AbilityCost.Loyalty is a POINTER: zero is a
// real printed cost that spends the turn's once-per-turn window, and
// an int could not tell it from "not a loyalty ability at all". This
// is the first card in the catalog to print one, so it is the first
// to exercise that distinction end to end.
//
// THE −2 SHIPS WEAKER THAN PRINTED. The token is created with its
// printed size, colour, type and vigilance; it does NOT carry its
// own "whenever you draw a card, put a +1/+1 counter on this token"
// trigger. Every ability hook in this engine is keyed on an oracle
// ID (game.CatalogKey), and a token has none — which is exactly why
// CreateTokenCopy works so well for copies and does nothing at all
// for a token that is its own new object with printed text. Giving
// this one a fabricated oracle ID to hang a trigger on would put a
// card in the registry that is not a card. The body, the vigilance
// and the blocker are real; the growth is not.
//
// THE −12 IS NOT REGISTERED. It is two chained choices made by the
// opponent — pick a permanent to bounce, then shuffle away everything
// nonland that is left — and the engine has no "choose a permanent
// you control" prompt outside the sacrifice picker. Choosing for the
// opponent would make their decision for them, and a −12 that
// shuffled without the bounce would be a different ability. The
// loyalty still accrues past 12.
func init() {
	Register(Spec{
		OracleID:     "5f0fa785-37ab-46a8-9ec0-6f8b29576d93",
		Name:         "Teferi, Temporal Pilgrim",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"The Spirit token doesn't grow — it has vigilance, but not its own \"whenever you draw a card\" ability.",
			"The -12 isn't offered — making an opponent choose a permanent and shuffle the rest has no prompt yet.",
		},
		// Printed loyalty reaches the card through deck import
		// (ADR 0032 §1); this is the fallback for tokens, fixtures
		// and the dev spawner.
		StartingLoyalty: 4,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventDrawCard},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Actor == source.Controller
			},
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source,
					"Teferi, Temporal Pilgrim — put a loyalty counter on Teferi",
					teferiPilgrimLoyaltyCounter)
			},
		}},
		Activated: []ActivatedAbility{
			{
				Label: "[0]: Draw a card.",
				Cost:  LoyaltyCost(0),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return DrawCards{Player: item.Controller, N: 1}.Apply(NewContext(g, item))
				},
			},
			{
				Label: "−2: Create a 2/2 blue Spirit creature token with vigilance and \"Whenever you draw a card, put a +1/+1 counter on this token.\"",
				Cost:  LoyaltyCost(-2),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return CreateToken{
						Controller: item.Controller,
						Template:   teferiPilgrimSpiritToken(),
						N:          1,
					}.Apply(NewContext(g, item))
				},
			},
		},
	})
}

// teferiPilgrimLoyaltyCounter is the body of the draw trigger.
// Package-level so the triggered item captures nothing.
//
// The source may have left the battlefield between the trigger going
// on the stack and resolving — a draw in response to the −2, and a
// Teferi at 2 loyalty who paid it — and AddCounter does not gate on
// zone, so the guard is this file's job.
func teferiPilgrimLoyaltyCounter(g *game.Game, item *game.StackItem) error {
	if !b09SourceStillOnBattlefield(g, item) {
		return nil
	}
	return AddCounter{
		Target: item.SourceCardID,
		Kind:   game.CounterLoyalty,
		N:      1,
	}.Apply(NewContext(g, item))
}

// teferiPilgrimSpiritToken is the −2's 2/2 blue Spirit with
// vigilance. Local to this file rather than added to tokens.go: it
// is a shape no other card in the catalog makes, and the shared file
// is the one every concurrent card batch collides on.
func teferiPilgrimSpiritToken() game.Card {
	return game.Card{
		Name:      "Spirit",
		TypeLine:  "Token Creature — Spirit",
		Power:     2,
		Toughness: 2,
		Colors:    []string{"U"},
		Keywords:  []string{"vigilance"},
	}
}
