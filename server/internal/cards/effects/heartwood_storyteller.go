package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Heartwood Storyteller — 2/3 Creature — Treefolk for {1}{G}{G}
// (EDHREC rank 4017):
//
//	"Whenever a player casts a noncreature spell, each of that
//	 player's opponents may draw a card."
//
// A three-mana political creature: it does not stop anybody casting
// anything, it just makes the rest of the table richer when they do.
// In a green creature deck that casts almost nothing but creatures,
// the Storyteller is close to a one-sided Howling Mine.
//
// It is in the batch because the beneficiaries are computed from the
// CASTER, not from the Storyteller's controller, and no other card in
// the catalog does that. Read the two clauses carefully:
//
//   - "Whenever A PLAYER casts" — anyone at the table, the
//     Storyteller's own controller included.
//   - "each of THAT PLAYER's opponents" — the opponents of whoever
//     cast the spell.
//
// So when an opponent casts a Rampant Growth, the Storyteller's
// controller draws and so does everybody else except the caster. When
// the Storyteller's own controller casts a noncreature spell, they
// draw nothing and hand a card to every opponent — which is exactly
// the deckbuilding constraint the card is printed to impose.
//
// "Noncreature spell" is the spell's card types on the stack, so an
// artifact creature spell is a creature spell and gets nothing past
// it, while a Saga, an Equipment or a land-fetching sorcery all
// trigger it.
//
// "MAY draw" is a real prompt asked of each beneficiary
// independently, so a player who does not want the card — an empty
// library, a Sylvan Library tax, a hand-size concern — can decline
// while the others take theirs.
//
// One trigger per spell, one prompt per beneficiary. The caster is
// read at resolution off the item's carried trigger context
// (item.Trigger.Event.Actor, #1223) — by then the spell may have left
// the stack entirely, and the beneficiaries are defined relative to
// whoever cast it, not to what's still there.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "82e6da87-f8a4-4897-bf0f-c0f2cd06b8b1",
		Name:         "Heartwood Storyteller",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventCast},
			AppliesTo: func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
				spell, ok := g.LookupCardForEffect(ev.CardID)
				return ok && !spell.IsCreature()
			},
			Key: "Heartwood Storyteller — each of that player's opponents may draw a card",
			Effect: func(g *game.Game, item *game.StackItem) error {
				return b38EachOpponentOfMayDraw(g, item, item.Trigger.Event.Actor)
			},
		}},
	})
}

// b38EachOpponentOfMayDraw asks every seated player other than
// `caster` whether to draw a card — Heartwood Storyteller's payout.
// Declared here rather than in batch38_helpers.go because it is one
// card's clause, not a shape the batch shares.
func b38EachOpponentOfMayDraw(g *game.Game, item *game.StackItem, caster uuid.UUID) error {
	ctx := NewContext(g, item)
	for _, p := range g.Seats {
		if p == nil || p.Eliminated || p.ID == caster {
			continue
		}
		who := p.ID
		if err := (MayChoice{
			Player:   who,
			Question: "Heartwood Storyteller — draw a card?",
			OnYes: func(ctx *Context) error {
				return DrawCards{Player: who, N: 1}.Apply(ctx)
			},
		}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}
