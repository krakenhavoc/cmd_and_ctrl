package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Aquatic Incursion — Enchantment {3}{U} (Jumpstart 2022):
//
//	"When this enchantment enters, create two 1/1 blue Merfolk creature tokens
//	 with hexproof.
//	 {3}{U}: Target Merfolk can't be blocked this turn."
//
// Two hexproof Merfolk on entry; the activated ability is Daughter of the
// Deep's (any Merfolk, not only the controller's, a target gone by resolution
// is left alone, CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e4212b05-d397-49fb-af8b-9e52a60e6d1e",
		Name:         "Aquatic Incursion",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 2, CreatureTokens: 2},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Aquatic Incursion — create two 1/1 blue Merfolk creature tokens with hexproof",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 blue Merfolk with hexproof"), N: 2})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}{U}: Target Merfolk can't be blocked this turn.",
			Purpose: effects.ForTargets(effects.CantBeBlockedUntilEOT(0, 0, 0)),
			Cost:    effects.ManaCost("{3}{U}"),
			Targets: effects.TargetCreature("target Merfolk", effects.HasSubtype("Merfolk")),
			Effect:  effects.Msc17CantBeBlockedTheTarget("Aquatic Incursion — can't be blocked"),
		}},
	})
}
