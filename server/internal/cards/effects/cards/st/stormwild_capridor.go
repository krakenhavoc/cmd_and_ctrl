package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Stormwild Capridor — Creature — Bird Goat {2}{W}, 1/3:
//
//	"Flying
//	 If noncombat damage would be dealt to this creature, prevent that damage. Put a +1/+1 counter on this creature for each 1 damage prevented this way."
//
// ADR 0108 §8 (#1906): a prevention static whose additional effect counts
// the damage PREVENTED this way, so damage that can't be prevented puts
// on no counter (CR 615.12; the ruling). Combat damage is untouched. One
// application per recipient in a damage instance, with the instance's
// total, so the counters arrive before lethal damage is checked.
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "2e83e8c8-adc4-4e33-9f8b-966b7231a6c9",
		Name:            "Stormwild Capridor",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Replacements: []game.ReplacementEffect{
			effects.PreventDamageDealtTo(effects.PreventionStatic{
				To:     effects.ToThisCreature,
				Damage: effects.NoncombatDamage,
				Then:   effects.CountersOnThisPerPreventedBody,
				Label:  "Stormwild Capridor — prevent noncombat damage to it and put a +1/+1 counter on it for each 1 prevented",
			}),
		},
	})
}
