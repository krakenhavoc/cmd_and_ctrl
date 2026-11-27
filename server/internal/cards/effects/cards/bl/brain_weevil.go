package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Brain Weevil — Creature — Insect {3}{B}, 1/1 (Duel Decks: Izzet vs. Golgari):
//
//	"Intimidate (This creature can't be blocked except by artifact creatures
//	 and/or creatures that share a color with it.)
//	 Sacrifice this creature: Target player discards two cards. Activate only
//	 as a sorcery."
//
// Dementia Bat's sacrifice-for-discard with no mana and sorcery timing: the
// sacrifice is the cost, so the discard still happens if the Weevil is gone by
// resolution, and a player with fewer than two cards discards what they have
// (CR 701.9b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "dfa8de88-9222-435f-a56e-c32e91fd5289",
		Name:            "Brain Weevil",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"intimidate"},
		Activated: []effects.ActivatedAbility{{
			Label:        "Sacrifice this creature: Target player discards two cards. Activate only as a sorcery.",
			Cost:         effects.SacrificeThis(),
			SorcerySpeed: true,
			Targets:      effects.TargetPlayer("target player"),
			Effect:       effects.TargetOpponentDiscards(2),
		}},
	})
}
