package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

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
// THE −11 IS NOW REGISTERED (#1230). It was blocked on a missing
// engine seam rather than a judgement call: a library search could
// not send what it finds to EXILE (game.searchDestZoneLocked accepted
// hand, battlefield, library and graveyard only) and SearchLibrarySpec
// had no "any number" spelling (Limit <= 0 meant exactly 1). Both are
// parameters on the search primitive now — Dest: game.ZoneExile and
// Unbounded: true — routed through searchRoute /
// routeCardToZoneLocked exactly like every other search take, so
// CR 614 and CR 903.9 still see the move. The free-cast grant is the
// half that was already writable: game.CastPermission with a "{0}"
// Cost override and the zero Duration ("until end of turn"), handed
// every card the search actually exiled.
//
// Printed loyalty reaches the card through deck import (ADR 0032
// §1), so Spec.StartingLoyalty is deliberately not set here.
func init() {
	Register(Spec{
		OracleID:     "5c58353a-fd60-4528-bf0d-669626cda0b2",
		Name:         "Ugin, Eye of the Storms",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			{
				FromStack: true,
				Watches:   []game.EventKind{game.EventCast},
				AppliesTo: Self,
				Key:       "Ugin, Eye of the Storms — exile a colored permanent",
				Targets:   uginEyeOfTheStormsTarget(),
				// A fill-in Build (ADR 0041 P9): this trigger fires as
				// the spell is announced (FromStack), before the
				// permanent exists, so the controller is the caster
				// (ev.Actor) rather than the default source.Controller.
				Build: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
					item := game.NewTriggeredItem(source, "Ugin, Eye of the Storms — exile a colored permanent")
					item.Controller, item.Owner = ev.Actor, ev.Actor
					return item
				},
				Effect: b27ExileChosenTarget,
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
			{
				Label: "−11: Search your library for any number of colorless nonland cards, exile them, then shuffle. " +
					"Until end of turn, you may cast those cards without paying their mana costs.",
				Cost: LoyaltyCost(-11),
				Effect: func(g *game.Game, item *game.StackItem) error {
					controller := item.Controller
					source := item.SourceCardID
					return g.SearchLibraryThenForEffect(game.SearchLibrarySpec{
						Player: controller,
						Source: source,
						Reason: "Ugin, Eye of the Storms — search for colorless nonland cards",
						Pred: func(c game.Card) bool {
							return c.IsColorless() && !c.IsLand()
						},
						Dest:      game.ZoneExile,
						Unbounded: true,
						Shuffle:   true,
						Then: func(g *game.Game, found []uuid.UUID) error {
							if len(found) == 0 {
								return nil
							}
							cards := make([]game.Card, 0, len(found))
							for _, id := range found {
								if c, ok := g.LookupCardForEffect(id); ok {
									cards = append(cards, c)
								}
							}
							g.GrantCastPermissionToCardsForEffect(game.CastPermission{
								Player: controller,
								Zone:   game.ZoneExile,
								Cost:   "{0}",
								Source: source,
								Label:  "Ugin, Eye of the Storms — cast without paying its mana cost",
							}, cards)
							return nil
						},
					})
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
