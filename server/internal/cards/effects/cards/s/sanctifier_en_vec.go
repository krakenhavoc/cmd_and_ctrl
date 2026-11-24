package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sanctifier en-Vec — Creature — Human Cleric {W}{W}, 2/2 (Modern Horizons 2):
//
//	"Protection from black and from red
//	 When this creature enters, exile all cards that are black or red from all
//	 graveyards.
//	 If a black or red permanent, spell, or card not on the battlefield would be
//	 put into a graveyard, exile it instead."
//
// Protection is two printed keywords. The enters trigger exiles the matching
// cards of every graveyard at once. The replacement (CR 614) is the graveyard
// exile family narrowed to a black or red object, whatever zone it comes from,
// a permanent dying, a spell finishing or being countered, a discard or a
// mill; a colourless or other-coloured card goes to the graveyard as usual. The
// colour words are read through the Sanctifier's text (CR 612.2).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "810d3afe-c644-444f-a4d9-88b323f0b581",
		Name:            "Sanctifier en-Vec",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"protection from black", "protection from red"},
		Replacements:    []game.ReplacementEffect{effects.OldB035SanctifierReplacement()},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Sanctifier en-Vec — exile all cards that are black or red from all graveyards", effects.OldB035SanctifierExile),
		},
	})
}
