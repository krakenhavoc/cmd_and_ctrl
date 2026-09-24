package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// spirit_guide.go — #1228: a mana ability that functions from the
// HAND (CR 113.6).
//
//	Simian Spirit Guide  "Exile this card from your hand: Add {R}."
//	Elvish Spirit Guide  "Exile this card from your hand: Add {G}."
//
// The whole card is the ability. Both are ordinary creature cards
// that are never meant to be cast — a 2/2 Ape for {2}{R} and a 2/2
// Elf for {2}{G} — and every deck that runs one runs it as a ritual
// that is also a body of last resort.
//
// Rules-wise it is three sentences, all of them already written
// somewhere in the engine and none of them previously reachable from
// a mana ability:
//
//   - CR 605.1a makes "Add {R}" a mana ability: it could add mana, it
//     is not a loyalty ability and it targets nothing. So it resolves
//     immediately, with no stack and no priority window, and can be
//     activated any time the player could produce mana — including
//     while casting a spell (CR 601.2g).
//   - CR 113.6 is what lets it work at all from a hand. The ability
//     says where it functions ("from your hand"), which is the
//     saying the rule asks for; without the declaration a card in a
//     hand has no abilities that do anything.
//   - CR 108.4 makes the "you" its OWNER, because a card in a hand
//     has no controller.
//
// The cost is `ManaAbilityCost.ExileSelf`, which is the same clause
// `AbilityCost.ExileSelf` carries for scavenge and embalm (#1221) —
// one component with two owners, so the CR 903.9 answer a commander
// exiled this way gets and the CR 601.2h indivisible step are each
// written once. See game/exile_cost.go and game/mana_ability_zone.go.

// ExileFromHandForMana builds "Exile this card from your hand: Add
// <produced>" — the Spirit Guides' ability, and the constructor for
// any future printing of the same clause.
//
//	ManaAbilities: []ManaAbility{ExileFromHandForMana("{R}")},
//
// The zone, the cost and the label travel together on purpose. A
// declaration missing any one of them is refused at boot
// (checkManaAbilityZones): a hand zone with no exile cost is a free
// repeatable mana source, and an exile cost with no hand zone has
// nothing to exile from.
//
// `produced` is a Scryfall brace string, exactly as ManaAbility.
// Produced is everywhere else — "{R}", "{G}", and a pipe would work
// if anything ever printed a choice here.
func ExileFromHandForMana(produced string) ManaAbility {
	return ManaAbility{
		Label:    "Exile this card from your hand: Add " + produced,
		Produced: produced,
		Zones:    []game.ZoneKind{game.ZoneHand},
		Cost:     ManaAbilityCost{ExileSelf: true},
	}
}
