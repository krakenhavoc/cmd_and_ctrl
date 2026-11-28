package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dream Fighter — Creature — Human Soldier {2}{U}, 1/1 (Mirage, S82 OLD batch
// 810, #5633):
//
//	"Whenever this creature blocks or becomes blocked by a creature, this
//	 creature and that creature phase out. (While they're phased out, they're
//	 treated as though they don't exist. Each one phases in before its
//	 controller untaps during their next untap step.)"
//
// One trigger for each blocker-attacker pair that names the Fighter (CR 509.3b,
// 509.3d); both creatures phase out together as it resolves (CR 702.26), which
// takes them out of combat, so no combat damage is dealt between them.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ea8d1c55-e045-4a3f-92f0-4ecd86362b97",
		Name:         "Dream Fighter",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.S82OldB810PhaseOutWithBlockPairFoe("Dream Fighter — this creature and that creature phase out"),
		},
	})
}
