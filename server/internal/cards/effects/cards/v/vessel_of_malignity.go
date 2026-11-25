package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vessel of Malignity — Enchantment {1}{B} (Shadows over Innistrad, S82 OLD):
//
//	"{1}{B}, Sacrifice this enchantment: Target opponent exiles two cards from
//	 their hand. Activate only as a sorcery."
//
// Aim for the Head's exile-two, as an activated ability: the opponent picks
// their own two cards.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "795a9f45-e4ee-457c-b260-17b9f09c3ba6",
		Name:         "Vessel of Malignity",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:        "{1}{B}, Sacrifice this enchantment: Target opponent exiles two cards from their hand. Activate only as a sorcery.",
			Cost:         effects.Plus(effects.ManaCost("{1}{B}"), effects.SacrificeThis()),
			SorcerySpeed: true,
			Targets:      effects.TargetPlayer("target opponent", effects.Opponent()),
			Effect:       effects.OldB286OpponentExilesTwoFromHand,
		}},
	})
}
