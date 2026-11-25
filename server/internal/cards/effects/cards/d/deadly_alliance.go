package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Deadly Alliance — Instant {4}{B} (Zendikar Rising, #5508):
//
//	"This spell costs {1} less to cast for each creature in your party.
//	 (Your party consists of up to one each of Cleric, Rogue, Warrior, and
//	 Wizard.)
//	 Destroy target creature or planeswalker."
//
// The party (CR 700.8) is counted as the spell is cast; the destroy is
// Bitter Triumph's, skipped if the target left (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:          "0d2cdcc6-da65-4f66-ba38-111efad44919",
		Name:              "Deadly Alliance",
		Completeness:      effects.CompletenessFull,
		SelfCostModifiers: []game.CostModifier{effects.OldB095PartyCostsLess()},
		Targets:           effects.TargetPermanent("target creature or planeswalker", effects.Or(effects.Creature(), effects.Planeswalker())),
		OnResolve:         effects.DestroyTheTargetPermanent,
	})
}
