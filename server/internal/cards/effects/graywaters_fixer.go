package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Graywater's Fixer — Creature — Lizard Mercenary {2}{B}{R} 4/4 (Outlaws
// of Thunder Junction Commander, #5429):
//
//	"Each outlaw creature card in your graveyard has encore {X}, where X is
//	 its mana value. (Exile it and pay its encore cost: For each opponent,
//	 create a token copy that attacks that opponent this turn if able. They
//	 gain haste. Sacrifice them at the beginning of the next end step.
//	 Activate only as a sorcery.)"
//
// A graveyard grant (ADR 0093's amendment of 2026-10-13, Sedris's shape) of
// the shared encore row (Encore, CR 702.141a) with one change: the mana
// component is ItsManaValue (game.AbilityCost.ManaValueFromSource), generic
// mana equal to the mana value of the card that has the row, written into
// the row as it is built. So one bundle prices a {B} Rogue at {1} and a
// {3}{R}{R} Pirate at {5}; an {X} in the card's mana cost is 0 in a
// graveyard (CR 202.3e) and a card with no mana cost has encore {0}.
//
// An outlaw is an Assassin, Mercenary, Pirate, Rogue or Warlock
// (CR 700.12), read off the card in the graveyard, so a changeling card is
// one. The row is the card's own ability: it leaves with the Fixer or when
// the Fixer loses its abilities, and an encore already activated resolves
// regardless (CR 113.7a). A card that prints encore keeps its own row
// beside the granted one.
//
// No simplification.
func init() {
	encore := Encore("{X}")
	encore.Label = "Encore {X}, where X is its mana value ({X}, Exile this card from your graveyard: For each opponent, " +
		"create a token that's a copy of it that attacks that opponent this turn if able. They gain haste. " +
		"Sacrifice them at the beginning of the next end step. Activate only as a sorcery.)"
	encore.Cost = Plus(ItsManaValue(), ExileThis())
	Register(Spec{
		OracleID:     "93438939-3909-4e59-bfe4-760ec73eb57d",
		Name:         "Graywater's Fixer",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key:       graywatersFixerEncore,
			Zone:      game.ZoneGraveyard,
			Activated: []ActivatedAbility{encore},
			Text:      "Encore {X}, where X is its mana value",
		}},
		Static: []game.StaticAbility{
			Labelled(GrantAbilitiesInZone(game.ZoneGraveyard, graywatersFixerRecipient, graywatersFixerEncore),
				"Each outlaw creature card in your graveyard has encore {X}, where X is its mana value."),
		},
	})
}

// graywatersFixerEncore is the bundle key of the encore the Fixer gives. On
// disk.
const graywatersFixerEncore = "graywaters-fixer/encore"

// graywatersFixerRecipient is "each outlaw creature card in your
// graveyard": a creature card its controller owns with an outlaw creature
// type. "Outlaw" is a word of the Fixer's text only as a batching term, so
// there is no type word for a text-changing effect to rewrite. A card whose
// mana cost the engine cannot read into a mana value is given nothing
// rather than an encore priced at {0}; no creature card prints such a cost.
func graywatersFixerRecipient(card *game.Card, g *game.Game, source *game.Card) bool {
	if source == nil || card.Owner != source.Controller || !card.IsCreature() {
		return false
	}
	if _, ok := card.ParsedManaValue(); !ok {
		return false
	}
	return outlaw()(g, source.Controller, *card)
}

// ItsManaValue is "{X}, where X is its mana value" as the mana component of
// a granted ability's cost (game.AbilityCost.ManaValueFromSource):
// Graywater's Fixer's encore. The bundle row names no mana of its own; the
// card that has the row supplies generic mana equal to its mana value
// (CR 202.3) as the row is built. ItsManaCost's sibling.
func ItsManaValue() game.AbilityCost { return game.AbilityCost{ManaValueFromSource: true} }
