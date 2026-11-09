package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Naktamun — Plane — Amonkhet (Commander Masters, #4870):
//
//	"Each creature card in your graveyard has embalm. Its embalm cost is
//	 equal to its mana cost. (Exile a creature card from your graveyard and
//	 pay its embalm cost: Create a token that's a copy of it, except it's a
//	 white Zombie in addition to its other types with no mana cost. Embalm
//	 only as a sorcery.)
//	 Whenever chaos ensues, you may discard a card. If you do, draw a card."
//
// A plane (ADR 0216 §6). The static is a zone grant (ADR 0093): each creature
// card the planar controller owns in a graveyard gets the embalm ability,
// whose cost is that card's own mana cost (game.AbilityCost.
// ManaIsSourceManaCost, stamped as the granted row is built) and which is
// activated from the graveyard as a sorcery; the planar controller activates
// it as its owner, per the ruling about a creature card put into the graveyard
// in the main phase. A card with no mana cost has no embalm cost to pay (CR
// 118.6). The chaos ability (CR 311.7) is a "you may discard a card"; only a
// discard draws ("if you do").
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a2048704-0d22-435d-b819-269291f2981e",
		Name:         "Naktamun",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key:       naktamunEmbalmGrant,
			Zone:      game.ZoneGraveyard,
			Activated: []ActivatedAbility{naktamunEmbalm()},
			Text:      "Embalm (equal to its mana cost)",
		}},
		Static: []game.StaticAbility{
			Labelled(GrantAbilitiesInZone(game.ZoneGraveyard, CardsYouOwn(func(c *game.Card) bool { return c.IsCreature() }), naktamunEmbalmGrant),
				"Each creature card in your graveyard has embalm. Its embalm cost is equal to its mana cost."),
		},
		Triggered: []game.TriggeredAbility{
			WheneverChaosEnsues("Naktamun — you may discard a card; if you do, draw a card",
				naktamunChaos),
		},
	})
}

// naktamunEmbalmGrant is the granted bundle's key. An on-disk key.
const naktamunEmbalmGrant = "naktamun/embalm"

// naktamunEmbalm is Embalm with its cost "equal to its mana cost".
func naktamunEmbalm() ActivatedAbility {
	ab := Embalm("")
	ab.Label = "Embalm — its mana cost (Exile this card from your graveyard: Create a token that's a copy of it, " +
		"except it's white, it has no mana cost, and it's a Zombie in addition to its other types. Activate only as a sorcery.)"
	ab.Cost = Plus(ManaCostOfThis(), ExileThis())
	return ab
}

// naktamunChaos is "you may discard a card. If you do, draw a card".
func naktamunChaos(g *game.Game, item *game.StackItem) error {
	return b39MayDiscardThenDraw(1, false, "Naktamun — you may discard a card; if you do, draw a card", drawOneIfDiscarded)(NewContext(g, item))
}
