package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ugin, Eye of the Storms — Legendary Planeswalker — Ugin {7},
// starting loyalty 7:
//
//	"When you cast this spell, exile up to one target permanent
//	 that's one or more colors.
//	 Whenever you cast a colorless spell, exile up to one target
//	 permanent that's one or more colors.
//	 +2: You gain 3 life and draw a card.
//	 0: Add {C}{C}{C}.
//	 −11: Search your library for any number of colorless nonland
//	      cards, exile them, then shuffle. Until end of turn, you may
//	      cast those cards without paying their mana costs."
//
// The colourless payoff: a seven-drop that answers a permanent as it
// is cast and then answers one more every time you cast anything
// colourless — which, in the deck that plays him, is most turns.
//
// THE TWO REMOVAL TRIGGERS ARE THE CARD AND BOTH ARE COMPLETE. The
// cast trigger is a FromStack ability (cascade's mechanism, Kozilek's
// and Oblivion Sower's shape): it fires as the spell is announced,
// takes its target then, and resolves ABOVE Ugin, so a countered
// Ugin has still exiled something. The second is an ordinary
// battlefield trigger, and it deliberately does NOT fire on Ugin's
// own cast — Ugin is not on the battlefield while he is on the
// stack, so the ability does not exist yet, which is exactly why the
// card prints the first trigger separately.
//
// "Up to one target" is a zero-floor target clause, so both triggers
// are fine with an empty board and with a table of nothing but
// colourless permanents — nothing is exiled and nothing errors.
//
// THE 0 IS NOT A MANA ABILITY. CR 605.1a excludes loyalty abilities
// from the mana-ability definition, so "0: Add {C}{C}{C}" uses the
// stack and can be responded to. It is therefore an ActivatedAbility
// applying the AddMana primitive, not an entry in ManaAbilities —
// putting it there would make it uncounterable and let it be
// activated at instant speed.
//
// THE −11 IS NOT REGISTERED, and the reason is a missing engine
// seam rather than a judgement call. A library search cannot send
// what it finds to EXILE: game.searchDestZoneLocked accepts hand,
// battlefield, library and graveyard, and returns ErrZoneNotFound for
// anything else, which would make the whole ability — search,
// shuffle and all — fail on its first line. The free-cast grant on
// the other side of the sentence is writable (game.CastPermission
// with a "{0}" override); the search that feeds it is not. Routing
// it through the hand instead would put every card the search found
// in a hand it never legally occupied, where a discard, a hand-size
// check or an opponent's Thoughtseize could see it — a different
// ability with different answers. So the whole clause is omitted,
// which leaves Ugin weaker than printed (#259's direction) and
// leaves him a plus, so he can still be used without only ticking
// down. Recorded in docs/engine-seams.md.
//
// Printed loyalty reaches the card through deck import (ADR 0032
// §1), so Spec.StartingLoyalty is deliberately not set here.
func init() {
	Register(Spec{
		OracleID:     "5c58353a-fd60-4528-bf0d-669626cda0b2",
		Name:         "Ugin, Eye of the Storms",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"The −11 isn't offered — Ugin can't search your library for colorless cards and set them aside to play.",
		},
		Triggered: []game.TriggeredAbility{
			{
				FromStack: true,
				Watches:   []game.EventKind{game.EventCast},
				AppliesTo: Self,
				Targets:   uginEyeOfTheStormsTarget(),
				Build: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
					return &game.StackItem{
						Kind:         game.StackItemTriggered,
						Controller:   ev.Actor,
						Owner:        ev.Actor,
						SourceCardID: source.InstanceID,
						Label:        "Ugin, Eye of the Storms — exile a colored permanent",
						Effect:       b27ExileChosenTarget,
					}
				},
			},
			Targeting(
				WheneverYouCast(Colorless(),
					"Ugin, Eye of the Storms — exile a colored permanent",
					b27ExileChosenTarget),
				uginEyeOfTheStormsTarget()),
		},
		Activated: []ActivatedAbility{
			{
				Label: "+2: You gain 3 life and draw a card.",
				Cost:  LoyaltyCost(2),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					if err := (GainLife{Player: item.Controller, Amount: 3}).Apply(ctx); err != nil {
						return err
					}
					return DrawCards{Player: item.Controller, N: 1}.Apply(ctx)
				},
			},
			{
				Label: "0: Add {C}{C}{C}.",
				Cost:  LoyaltyCost(0),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return AddMana{Player: item.Controller, Produced: "{C}{C}{C}"}.Apply(NewContext(g, item))
				},
			},
		},
	})
}

// uginEyeOfTheStormsTarget is the clause both removal triggers
// print, word for word. Built per trigger rather than shared as one
// pointer, so neither can mutate the other's spec.
func uginEyeOfTheStormsTarget() *game.TargetSpec {
	return TargetPermanent("up to one target permanent that's one or more colors", Not(Colorless())).WithCount(0, 1)
}
