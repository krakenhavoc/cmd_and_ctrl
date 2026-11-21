package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Young Deathclaws — Creature — Lizard Mutant {2}{B}{G}, 4/2 (Fallout,
// #5948):
//
//	"Menace
//	 Each creature card in your graveyard has scavenge. The scavenge
//	 cost is equal to its mana cost. (Exile a creature card from your
//	 graveyard and pay its mana cost: Put a number of +1/+1 counters
//	 equal to that card's power on target creature. Scavenge only as a
//	 sorcery.)"
//
// A graveyard grant (Graywater's Fixer's shape) of the shared scavenge
// row (Scavenge, CR 702.97a) with its mana component read off the card
// that has the row (ItsManaCost, game.AbilityCost.ManaFromSource, Cursecloth
// Wrappings'). The row is the card's own ability: it leaves with the
// Deathclaws or when they lose their abilities, and a scavenge already
// activated resolves regardless (CR 113.7a). A card that prints scavenge
// keeps its own row beside the granted one. A card with no mana cost would
// owe a cost it can't pay (CR 118.6), so it is given nothing.
//
// No simplification.
func init() {
	scavenge := effects.Scavenge("{0}")
	scavenge.Label = "Scavenge — its mana cost (Its mana cost, Exile this card from your graveyard: " +
		"Put a number of +1/+1 counters equal to this card's power on target creature. " +
		"Activate only as a sorcery.)"
	scavenge.Cost = effects.Plus(effects.ItsManaCost(), effects.ExileThis())
	effects.Register(effects.Spec{
		OracleID:        "f24fc213-6341-482c-a056-946d0f1dd76b",
		Name:            "Young Deathclaws",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Grants: []effects.AbilityGrant{{
			Key:       youngDeathclawsScavenge,
			Zone:      game.ZoneGraveyard,
			Activated: []effects.ActivatedAbility{scavenge},
			Text:      "Scavenge. The scavenge cost is equal to its mana cost.",
		}},
		Static: []game.StaticAbility{
			effects.Labelled(effects.GrantAbilitiesInZone(game.ZoneGraveyard, youngDeathclawsRecipient, youngDeathclawsScavenge),
				"Each creature card in your graveyard has scavenge. The scavenge cost is equal to its mana cost."),
		},
	})
}

// youngDeathclawsScavenge is the bundle key of the scavenge the Deathclaws
// give. On disk.
const youngDeathclawsScavenge = "young-deathclaws/scavenge"

// youngDeathclawsRecipient is "each creature card in your graveyard": a
// creature card its controller owns, with a mana cost to pay.
func youngDeathclawsRecipient(card *game.Card, _ *game.Game, source *game.Card) bool {
	return source != nil && card.Owner == source.Controller && card.IsCreature() && card.ManaCost != ""
}
