package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Falkenrath Gorger — Creature — Vampire Berserker {R}, 2/1:
//
//	"Each Vampire creature card you own that isn't on the battlefield has
//	 madness. The madness cost is equal to its mana cost. (If you discard a
//	 card with madness, discard it into exile. When you do, cast it for its
//	 madness cost or put it into your graveyard.)"
//
// Madness (CR 702.35a) matters only to a card discarded from a hand, so the
// Gorger carries madness's two halves for the cards it grants it to: the
// replacement that exiles the discarded Vampire instead of putting it into
// the graveyard, and the trigger that offers the cast for its mana cost
// (game/madness_grant.go). "You own" is the card's owner, so it works on a
// discard an opponent forces. The Vampire is read on its printed card (a
// card in a hand has no layers), by the word of the Gorger's own text
// (CR 612.2).
//
// No simplification.
func init() {
	grants := func(card, source *game.Card) bool {
		return card.Owner == source.Controller && card.IsCreature() && card.HasSubtype(source.TextSubtype("Vampire"))
	}
	Register(Spec{
		OracleID:     "26aba6fa-830f-4a7f-8618-488d8e00215a",
		Name:         "Falkenrath Gorger",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{game.GrantedMadnessReplacement(grants)},
		Triggered:    []game.TriggeredAbility{game.GrantedMadnessTrigger(grants)},
	})
}
