package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sanguinary Priest — Creature — Astartes Cleric {3}{B}, 2/4 (Warhammer 40,000
// Commander, #5948):
//
//	"Lifelink
//	 Blood Chalice — Whenever another creature you control dies, this creature
//	 deals 1 damage to any target."
//
// Elas il-Kor's "another creature you control dies" with a targeted body: the
// target is chosen as the ability goes on the stack (CR 603.3d) and the damage
// is dealt by the Priest, as it last existed if it has since left (CR 608.2h).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "22f5f39d-0898-4c47-8fa8-7d97102d3229",
		Name:            "Sanguinary Priest",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.Targeting(effects.On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.AnotherCreatureYouControlDied(ev, source, g)
			}, "Sanguinary Priest — 1 damage to any target", effects.DealDamageToTheTarget(1)), effects.TargetAny()),
				game.Purpose{DeathPayoff: true}),
		},
	})
}
