package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Alms Beast — Creature — Beast {2}{W}{B}, 6/6 (Gatecrash, S82 OLD b408,
// #5565):
//
//	"Creatures blocking or blocked by this creature have lifelink."
//
// A layer 6 keyword grant (KeywordGrant) to every creature that is blocking the
// Beast or that the Beast is blocking, read live, so the creature has lifelink
// while the combat lasts and loses it when the combat ends or the Beast leaves.
// It reads combat status, so it declares DependsOnAttackingStatus and
// DependsOnEvents{EventBlock}: the layer cache is dropped when attackers or
// blockers are declared and when combat ends.
//
// No simplification.
func init() {
	grant := effects.KeywordGrant(func(target *game.Card, _ *game.Game, source *game.Card) bool {
		if target == nil || source == nil || target.InstanceID == source.InstanceID || !target.IsCreature() {
			return false
		}
		blocking := target.IsBlockingAttacker(source.InstanceID)
		blocked := source.IsBlockingAttacker(target.InstanceID)
		return blocking || blocked
	}, "lifelink")
	grant.Label = "Creatures blocking or blocked by this creature have lifelink."
	grant.DependsOnAttackingStatus = true
	grant.DependsOnEvents = []game.EventKind{game.EventBlock}
	effects.Register(effects.Spec{
		OracleID:     "b221939c-93e7-4ffc-b995-5b67e173ee15",
		Name:         "Alms Beast",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{grant},
	})
}
