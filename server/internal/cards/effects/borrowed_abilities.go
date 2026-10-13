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
	return GrantActivatedAbilitiesOf(func(target *game.Card, _ *game.Game, source *game.Card) bool {
		return source != nil && target.InstanceID == source.InstanceID
	}, b)
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
