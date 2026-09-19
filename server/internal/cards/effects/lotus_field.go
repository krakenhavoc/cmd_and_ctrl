package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Lotus Field — Land (#332):
//
//	"Hexproof
//	 This land enters tapped.
//	 When this land enters, sacrifice two lands.
//	 {T}: Add three mana of any one color."
//
// Hexproof is a printed keyword the deck importer stamps. The tapped
// entry is SelfEntersTapped. The enters trigger asks for two sacrifice
// prompts over the controller's lands, one land each (the Planar
// Engineering shape); Lotus Field itself is a legal choice, as the
// printed card allows, and a controller with fewer than two lands
// sacrifices what they have. The mana ability is #742's one-pick,
// three-token "any one color" (see Gilded Lotus).
//
// Declared difference: the printed card sacrifices both lands at once
// (one CR 701.21 event); the two prompts here sacrifice them one after
// the other. The same two lands end up in the graveyard, and the
// second prompt cannot name the first land (ResolveSacrificeChoice
// refuses a card no longer on the battlefield), so the difference is
// only in batching: the lands leave in two events instead of one. A
// "whenever one or more" watcher still fires once (both sacrifices
// fall in one batch — nothing resolves and no step begins between
// them, see AGENTS.md §7 — so OncePerBatch declines the second),
// and a per-land watcher fires twice either way. Neither direction
// is stronger than printed; it is declared because it is not the
// printed timing. The multi-select sacrifice-N picker (#747) removes it.
func init() {
	Register(Spec{
		OracleID:     "134d5b82-7940-4b33-a922-7f9d1f403e50",
		Name:         "Lotus Field",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"The two lands are sacrificed one at a time, through two prompts, rather than at the same moment."},
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Lotus Field — sacrifice two lands", func(g *game.Game, item *game.StackItem) error {
				g.PlayerSacrificesNForEffect(item.SourceCardID, item.Controller,
					sacrificeSpec("a land", Land()), "Lotus Field — sacrifice a land", 2)
				return nil
			}),
		},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: OneColorOfAmount(3),
			Label:    "Add three mana of any one color",
		}},
	})
}
