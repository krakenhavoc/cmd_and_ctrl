package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sliver Gravemother — Legendary Creature — Sliver {W}{U}{B}{R}{G}, 6/6
// (Commander Masters):
//
//	"The "legend rule" doesn't apply to Slivers you control.
//	 Each Sliver creature card in your graveyard has encore {X}, where X is
//	 its mana value.
//	 Encore {5} ({5}, Exile this card from your graveyard: For each
//	 opponent, create a token copy that attacks that opponent this turn if
//	 able. They gain haste. Sacrifice them at the beginning of the next end
//	 step. Activate only as a sorcery.)"
//
// Three pieces, each on a seam that already exists: the legend-rule
// exemption for your Slivers (Spider-Verse's shape, CR 704.5j), Graywater's
// Fixer's graveyard grant of an encore priced at the card's own mana value
// (CR 702.141a), and the shared encore row for the Gravemother's own {5}.
// The grant is the card's own ability, so it ends when the Gravemother
// leaves; an encore already activated resolves regardless (CR 113.7a). A
// Sliver that prints encore keeps its own row beside the granted one.
//
// No simplification.
func init() {
	encore := Encore("{X}")
	encore.Label = "Encore {X}, where X is its mana value ({X}, Exile this card from your graveyard: For each opponent, " +
		"create a token that's a copy of it that attacks that opponent this turn if able. They gain haste. " +
		"Sacrifice them at the beginning of the next end step. Activate only as a sorcery.)"
	encore.Cost = Plus(ItsManaValue(), ExileThis())
	Register(Spec{
		OracleID:             "0d59f010-9669-4ad2-9ae9-cd88b85f282c",
		Name:                 "Sliver Gravemother",
		Completeness:         CompletenessFull,
		LegendRuleExemptions: LegendRuleDoesntApplyToYourSubtype("Sliver", "Slivers"),
		Activated:            []ActivatedAbility{Encore("{5}")},
		Grants: []AbilityGrant{{
			Key:       sliverGravemotherEncore,
			Zone:      game.ZoneGraveyard,
			Activated: []ActivatedAbility{encore},
			Text:      "Encore {X}, where X is its mana value",
		}},
		Static: []game.StaticAbility{
			Labelled(GrantAbilitiesInZone(game.ZoneGraveyard, sliverGravemotherRecipient, sliverGravemotherEncore),
				"Each Sliver creature card in your graveyard has encore {X}, where X is its mana value."),
		},
	})
}

// sliverGravemotherEncore is the bundle key of the encore the Gravemother
// gives. On disk.
const sliverGravemotherEncore = "sliver-gravemother/encore"

// sliverGravemotherRecipient is "each Sliver creature card in your
// graveyard": a creature card its controller owns with the Sliver type
// (the word read through the Gravemother's text), changelings included.
func sliverGravemotherRecipient(card *game.Card, _ *game.Game, source *game.Card) bool {
	if source == nil || card.Owner != source.Controller || !card.IsCreature() {
		return false
	}
	if _, ok := card.ParsedManaValue(); !ok {
		return false
	}
	return card.HasSubtype(source.TextSubtype("Sliver"))
}
