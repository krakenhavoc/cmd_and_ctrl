package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sword of Sinew and Steel — Artifact — Equipment {3} (Mirrodin Besieged
// reprint, #5908):
//
//	"Equipped creature gets +2/+2 and has protection from black and from
//	 red.
//	 Whenever equipped creature deals combat damage to a player, destroy
//	 up to one target planeswalker and up to one target artifact.
//	 Equip {2}"
//
// Sword of Body and Mind's shape with Boom Box's two "up to one" clauses
// and resolution: the targets still legal are destroyed in one
// simultaneous event, so an artifact planeswalker chosen for both is
// destroyed once.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ccab4509-f189-4610-9597-547e7f6b0775",
		Name:         "Sword of Sinew and Steel",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			PumpAttached(2, 2),
			GrantToAttached("protection from black", "protection from red"),
		},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventDealDamage},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return attachedCreatureDealtCombatDamageToPlayer(ev, source, g)
			},
			Targets: Clauses(
				TargetPermanent("up to one target planeswalker", Planeswalker()).WithCount(0, 1),
				TargetPermanent("up to one target artifact", Artifact()).WithCount(0, 1),
			),
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Key:     "Sword of Sinew and Steel — destroy up to one target planeswalker and up to one target artifact",
			Effect: func(g *game.Game, item *game.StackItem) error {
				return destroyEachLegalTarget(item, NewContext(g, item))
			},
		}},
		Activated: []ActivatedAbility{
			EquipAbility("{2}"),
		},
	})
}
