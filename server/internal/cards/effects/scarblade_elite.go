package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Scarblade Elite — Creature — Elf Assassin {B}{B}, 2/2 (Morningtide,
// #5592):
//
//	"{T}, Exile an Assassin card from your graveyard: Destroy target
//	 creature."
//
// The exile is Moorland Haunt's graveyard cost (ExileFromGraveyard), the
// card named at announce (CR 602.2b); "an Assassin card" reads the card's
// subtypes, so a changeling card pays it. The body is District Mascot's
// destroy of the target still legal at resolution.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e5fbd5c8-7b0c-4740-9430-456e86267966",
		Name:         "Scarblade Elite",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label: "{T}, Exile an Assassin card from your graveyard: Destroy target creature.",
			Cost: Plus(TapCost(), ExileFromGraveyard(1, "an Assassin card", func(c game.Card) bool {
				return c.HasSubtype("Assassin")
			})),
			Targets: TargetCreature("target creature"),
			Effect:  destroyFirstLegalTarget,
		}},
	})
}
