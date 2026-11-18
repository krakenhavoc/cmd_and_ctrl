package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Network Terminal — Artifact {3} (Kamigawa: Neon Dynasty, S82):
//
//	"{T}: Add one mana of any color.
//	 {1}, {T}, Tap another untapped artifact you control: Draw a card, then
//	 discard a card."
//
// Citanul Stalwart's tap-another cost (TapAnotherUntapped: not the {T} symbol,
// so a summoning-sick artifact creature can be tapped for it, CR 302.6) and
// Adventurer's Airship's loot body (discard chosen after the draw).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a0eb1d5b-2c54-44fa-b7ad-bdd4b1147929",
		Name:         "Network Terminal",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{W|U|B|R|G}",
			Label:    "{T}: Add one mana of any color.",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{1}, {T}, Tap another untapped artifact you control: Draw a card, then discard a card.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{1}"), TapCost(), TapAnotherUntapped("another untapped artifact you control", Artifact())),
			Effect:  func(g *game.Game, item *game.StackItem) error { return b16DrawThenDiscard(g, item, 1) },
		}},
	})
}
