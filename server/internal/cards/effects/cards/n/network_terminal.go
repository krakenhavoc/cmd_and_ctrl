package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "a0eb1d5b-2c54-44fa-b7ad-bdd4b1147929",
		Name:         "Network Terminal",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{W|U|B|R|G}",
			Label:    "{T}: Add one mana of any color.",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, {T}, Tap another untapped artifact you control: Draw a card, then discard a card.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{1}"), effects.TapCost(), effects.TapAnotherUntapped("another untapped artifact you control", effects.Artifact())),
			Effect:  func(g *game.Game, item *game.StackItem) error { return effects.B16DrawThenDiscard(g, item, 1) },
		}},
	})
}
