package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Timeless Lotus — Legendary Artifact {5} (EDHREC rank 2181):
//
//	"Timeless Lotus enters tapped.
//	 {T}: Add {W}{U}{B}{R}{G}."
//
// The five-colour rock. Enters tapped is the CR 614 self-replacement
// (no tap event, as with every tapped land); the mana ability is
// five fixed single-colour slots, so every activation drops one of
// each colour into the pool with no picker — the printed "add
// {W}{U}{B}{R}{G}", not "add five mana of any colour", so nothing is
// narrowed to a commander's identity. Mana of a colour outside the
// identity is still produced, as printed; the identity rule is a
// deck-construction rule, not a mana rule.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "fec77e0a-433e-4d32-9260-e26a37f6ad23",
		Name:         "Timeless Lotus",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{W}{U}{B}{R}{G}",
			Label:    "Add {W}{U}{B}{R}{G}",
		}},
	})
}
