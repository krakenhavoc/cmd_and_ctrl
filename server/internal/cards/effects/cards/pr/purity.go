package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Purity — Creature — Elemental Incarnation {3}{W}{W}{W}, 6/6:
//
//	"Flying
//	 If noncombat damage would be dealt to you, prevent that damage. You
//	 gain life equal to the damage prevented this way.
//	 When Purity is put into a graveyard from anywhere, shuffle it into
//	 its owner's library."
//
// The prevention is an ADR 0108 §8 static (#1906) on noncombat damage
// to its controller, once per application, gaining the life prevented
// (the source shields' "prevention/gain-life-equal" body: the same
// sentence). Damage that can't be prevented is dealt and gains nothing
// (CR 615.12).
//
// The last line is the graveyard trigger (#2038,
// WhenThisIsPutIntoAGraveyardFromAnywhereShuffleItIn).
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "23ef92b0-2298-4e8d-bf64-46d096c84a9d",
		Name:            "Purity",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Replacements: []game.ReplacementEffect{
			effects.PreventDamageDealtTo(effects.PreventionStatic{
				To:     effects.ToYou,
				Damage: effects.NoncombatDamage,
				Then:   effects.PreventedGainLifeBody,
				Label:  "Purity — prevent noncombat damage to you and gain that much life",
			}),
		},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisIsPutIntoAGraveyardFromAnywhereShuffleItIn("Purity — shuffle it into its owner's library"),
		},
	})
}
