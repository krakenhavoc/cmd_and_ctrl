package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// granted_evoke.go — #5428: evoke a permanent gives the spells its
// controller casts. Ashling, the Limitless: "Elemental permanent spells
// you cast from your hand gain evoke {4} as you cast them."
//
// CR 702.74a: "Evoke represents two abilities: a static ability that
// functions in any zone from which the card with evoke can be cast and a
// triggered ability that functions on the battlefield. "Evoke [cost]"
// means "You may cast this card by paying [cost] rather than paying its
// mana cost" and "When this permanent enters, if its evoke cost was paid,
// its controller sacrifices it.""
//
// The first half is a granted alternative cost (ADR 0118 §3,
// granted_alternative_cost.go). The second is what no other granted offer
// but warp carries: a clause on the permanent the spell becomes. The
// spell gained evoke as it was cast (CR 610.5: "Some static abilities
// create one-shot effects that cause spells a player casts to gain an
// ability as that player casts them"), so it keeps it whatever becomes
// of the permanent that gave it, and the engine reads the sacrifice off
// the claim's key (game.GrantedAltCostEvoke), not off the battlefield.

// GrantedAltCostEvoke is the key of a granted evoke offer. It is
// game.GrantedAltCostEvoke: the entry hook reads it. An on-disk identity.
const GrantedAltCostEvoke = game.GrantedAltCostEvoke

// AltCostKeyEvoke is the key of a card's own evoke offer (Evoke,
// EvokePitch): an on-disk identity.
const AltCostKeyEvoke = "evoke"

// PaidEvoke reports whether `key` — a StackItem's or a permanent's
// AltCost — is an evoke cost: the card's own ("evoke") or one a permanent
// gave it ("granted-evoke"). "If its evoke cost was paid" is true for
// either, since a spell given evoke has that evoke cost (CR 702.74a).
func PaidEvoke(key string) bool {
	return key == AltCostKeyEvoke || key == GrantedAltCostEvoke
}

// EvokeForSpellsYouCastFromHand is "<spells> you cast from your hand
// gain evoke {cost} as you cast them" (Ashling, the Limitless): a granted
// alternative cost that reaches only casts from the hand, only of the
// spells `spells` names (read off the face being cast), and that carries
// evoke's sacrifice (CR 702.74a). A card that prints its own evoke keeps
// it as a second offer, and the caster picks one (CR 118.9a). `spellsText`
// words the spells for the ability row: "Elemental permanent spells".
func EvokeForSpellsYouCastFromHand(cost, spellsText string, spells game.PermissionFilter) game.GrantedAlternativeCost {
	return game.GrantedAlternativeCost{
		Offer: game.AlternativeCost{
			Key:              GrantedAltCostEvoke,
			Label:            "Evoke " + cost,
			ManaCost:         cost,
			SacrificeOnEntry: true,
		},
		Zones:      []game.ZoneKind{game.ZoneHand},
		Spells:     spells,
		SpellsText: spellsText,
	}
}
