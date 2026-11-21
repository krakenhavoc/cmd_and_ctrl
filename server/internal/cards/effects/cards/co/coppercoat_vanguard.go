package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Coppercoat Vanguard — Creature — Human Soldier {1}{W}, 2/2
// (March of the Machine: The Aftermath):
//
//	"Each other Human you control gets +1/+0 and has ward {1}."
//
// A tribal anthem (TribalAnthem, effective subtypes) for the +1/+0, and
// the ward is Bronze Guardian's WardGranted: ward is a triggered
// ability (CR 702.21a), so the grant is the Vanguard carrying the ward
// trigger for every Human it covers, judged with the same filter as the
// anthem.
//
// No simplification.
func init() {
	humans := effects.TribeFilter{Tribes: []string{"Human"}, Others: true, YoursOnly: true}
	effects.Register(effects.Spec{
		OracleID:     "9f6fc734-4935-42d9-93a0-1b5ad8d8d277",
		Name:         "Coppercoat Vanguard",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.TribalAnthem(humans, 1, 0),
		},
		Triggered: []game.TriggeredAbility{
			effects.WardGranted(effects.WardMana("{1}"), "Coppercoat Vanguard — each other Human you control has ward {1}", humans.Matches),
		},
	})
}
