package effects

import (
	"fmt"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// borrowed_abilities.go — the card side of ADR 0155 (#1557): "has all
// activated abilities of …". The engine half, the layer-6 borrow step and
// the readers, is server/internal/game/borrowed_abilities.go.
//
// A borrow is a layer-6 static that declares game.StaticAbility.Borrows.
// Build one with HasActivatedAbilitiesOf ("~ has all activated abilities
// of …") or GrantActivatedAbilitiesOf ("[recipients] have all activated
// abilities of …"); the lender sets (battlefield, graveyard, exiled-with,
// library top, chosen permanent) arrive with the cards that use them.

// HasActivatedAbilitiesOf is "~ has all activated abilities of …": a
// layer-6 borrow whose one recipient is its own source.
func HasActivatedAbilitiesOf(b game.AbilityBorrow) game.StaticAbility {
	s := GrantActivatedAbilitiesOf(func(target *game.Card, _ *game.Game, source *game.Card) bool {
		return source != nil && target.InstanceID == source.InstanceID
	}, b)
	// The printed clause is the row's player-facing wording.
	s.Label = b.Label
	return s
}

// GrantActivatedAbilitiesOf is "[recipients] have all activated abilities
// of …" (Hazel's Brewmaster's Foods, Agatha's Soul Cauldron's creatures):
// a layer-6 borrow over the ordinary StaticAbility predicate.
func GrantActivatedAbilitiesOf(appliesTo func(target *game.Card, g *game.Game, source *game.Card) bool, b game.AbilityBorrow) game.StaticAbility {
	borrow := b
	return game.StaticAbility{
		Layer:     game.Layer6Ability,
		AppliesTo: appliesTo,
		Borrows:   &borrow,
		Label:     b.Label,
	}
}

// LenderPredicate picks the lenders of a borrow. It is a CardPredicate that
// also sees the borrowing source, for the clauses that compare with it
// ("that don't have the same name as this creature"). `source.Controller`
// is "you".
type LenderPredicate func(g *game.Game, source *game.Card, c game.Card) bool

// ByController lifts a CardPredicate (YouControl(), Creature() …) to a
// LenderPredicate whose caster is the borrowing source's controller.
func ByController(p CardPredicate) LenderPredicate {
	return func(g *game.Game, source *game.Card, c game.Card) bool {
		return p(g, source.Controller, c)
	}
}

// BattlefieldLenders is "… the activated abilities of [permanents matching
// P]": every permanent on the battlefield that matches. The borrowing
// source itself is never its own lender, because its own abilities are
// already its own ("each other creature", "creatures you control that
// don't have the same name"). Types, control and counters are read after
// layers 1-5 have settled, and the lender's rows after its own layer 6
// (ADR 0155 Decision 3).
func BattlefieldLenders(p LenderPredicate) func(g *game.Game, source *game.Card) []*game.Card {
	return func(g *game.Game, source *game.Card) []*game.Card {
		if g == nil || g.Battlefield == nil || source == nil {
			return nil
		}
		var out []*game.Card
		for i := range g.Battlefield.Cards {
			c := &g.Battlefield.Cards[i]
			if c.InstanceID == source.InstanceID || !p(g, source, *c) {
				continue
			}
			out = append(out, c)
		}
		return out
	}
}

// LenderGraveyards says whose graveyards a graveyard lender reads.
type LenderGraveyards uint8

const (
	// YourGraveyard is "cards in your graveyard" (Thranduil, Trazyn): the
	// borrowing source's controller's.
	YourGraveyard LenderGraveyards = iota
	// AllGraveyards is "cards in all graveyards" (Necrotic Ooze, Mirran
	// Safehouse).
	AllGraveyards
)

// GraveyardLenders is "… the activated abilities of [cards matching P] in
// [your / all] graveyards". A card there has no layered characteristic, so
// the engine reads its printed abilities through its catalog entry and its
// designation gate (CR 113.6 keeps an ability that works only on the
// battlefield from being lent as usable there, which the activation path
// enforces). The borrowing source is never its own lender.
func GraveyardLenders(whose LenderGraveyards, p LenderPredicate) func(g *game.Game, source *game.Card) []*game.Card {
	return func(g *game.Game, source *game.Card) []*game.Card {
		if g == nil || source == nil {
			return nil
		}
		var out []*game.Card
		for _, seat := range g.Seats {
			if seat == nil || seat.Graveyard == nil {
				continue
			}
			if whose == YourGraveyard && seat.ID != source.Controller {
				continue
			}
			for i := range seat.Graveyard.Cards {
				c := &seat.Graveyard.Cards[i]
				if c.InstanceID == source.InstanceID || !p(g, source, *c) {
					continue
				}
				out = append(out, c)
			}
		}
		return out
	}
}

// BorrowedAbilitiesSpendableAsAnyColor is the rider "You may spend mana as
// though it were mana of any color to activate those abilities." for the
// borrow whose printed clause is `borrowLabel` (Drana and Linvala): the
// borrower's controller, for the rows that borrow lent and no others
// (ADR 0155 Decision 7).
func BorrowedAbilitiesSpendableAsAnyColor(label, borrowLabel string) []game.AnyColorSpendStatic {
	return []game.AnyColorSpendStatic{{
		Label: label,
		Whose: game.AnyColorSpendYou,
		Covers: func(ctx game.ManaSpendContext) bool {
			return ctx.AbilityOrigin.LentByBorrow(borrowLabel)
		},
	}}
}

// checkStaticBorrows is ADR 0155 Decision 1's refusal. A borrow is an
// ability-adding effect, CR 613.1f's layer 6, and the engine applies it
// in its own step at the end of that bucket, so:
//
//   - declared in any other layer it would never be reached;
//   - it stands alone — a removal, a grant or an Apply beside it would be
//     applied in the borrow step's slot rather than its own timestamp's;
//   - it names its lenders and its printed clause, or it lends nothing
//     and shows nothing.
func checkStaticBorrows(card, what string, statics []game.StaticAbility) {
	for i, s := range statics {
		if s.Borrows == nil {
			continue
		}
		where := fmt.Sprintf("effects.Register: %q %s %d", card, what, i)
		switch {
		case s.Layer != game.Layer6Ability:
			panic(fmt.Sprintf("%s borrows abilities in layer %d — a borrow is a layer-6 effect (CR 613.1f)", where, s.Layer))
		case s.RemovesAbilities || len(s.GrantAbilities) > 0 || s.Apply != nil:
			panic(where + " borrows abilities beside a removal, a grant or an Apply — declare those as a static of their own")
		case s.AffectsSpells:
			panic(where + " borrows abilities for spells — a borrow lends to permanents")
		case s.AppliesTo == nil:
			panic(where + " borrows abilities for nothing — it has no AppliesTo")
		case s.Borrows.Lenders == nil:
			panic(where + " borrows abilities from nothing — it has no Lenders")
		case s.Borrows.Label == "":
			panic(where + " borrows abilities with no printed Label")
		}
	}
}
