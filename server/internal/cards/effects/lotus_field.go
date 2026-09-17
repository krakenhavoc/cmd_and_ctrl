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
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "134d5b82-7940-4b33-a922-7f9d1f403e50",
		Name:         "Lotus Field",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Lotus Field — sacrifice two lands", func(g *game.Game, item *game.StackItem) error {
				for i := 0; i < 2; i++ {
					if g.PlayerSacrificesForEffect(item.SourceCardID, item.Controller,
						sacrificeSpec("a land", Land()), "Lotus Field — sacrifice a land") == 0 {
						break
					}
				}
				return nil
			}),
		},
		ManaAbilities: []ManaAbility{{
			Cost:                    ManaAbilityCost{Tap: true},
			Produced:                OneColorOfAmount(3),
			Label:                   "Add three mana of any one color",
			IgnoreCommanderIdentity: true,
		}},
	})
}
