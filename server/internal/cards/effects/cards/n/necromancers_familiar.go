package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Necromancer's Familiar — Creature — Bird Spirit {3}{B}, 3/1 (Modern Horizons 2):
//
//	"Flying
//	 Hellbent — This creature has lifelink as long as you have no cards in hand.
//	 {B}, Discard a card: This creature gains indestructible until end of turn.
//	 Tap it."
//
// Hellbent is a layer 6 grant held to an empty hand and read live (so the
// discard that pays the activated ability can switch it on). The activated
// ability's discard is a cost; the grant and then the tap are one effect, and a
// Familiar that left the battlefield is neither protected nor tapped (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ce8833ea-5f58-4ab9-b33d-8972f2215574",
		Name:            "Necromancer's Familiar",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static:          []game.StaticAbility{effects.OldB034HellbentKeyword("lifelink")},
		Activated: []effects.ActivatedAbility{{
			Label:   "{B}, Discard a card: This creature gains indestructible until end of turn. Tap it.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.Plus(effects.ManaCost("{B}"), effects.DiscardACard()),
			Effect:  effects.OldB034GrantIndestructibleThenTap,
		}},
	})
}
