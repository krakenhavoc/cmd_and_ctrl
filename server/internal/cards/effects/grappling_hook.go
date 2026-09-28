package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Grappling Hook — Artifact — Equipment, {4}:
//
//	"Equipped creature has double strike.
//	 Whenever equipped creature attacks, you may have target creature
//	 block it this turn if able.
//	 Equip {4}"
//
// #1684: "it" is the EQUIPPED creature — the attacker the trigger's
// event names — so the requirement is a blocksAttacker record naming
// that attacking object (TargetBlocksTheAttacker). Unlike Provoke, the
// target is any creature and is not untapped: a tapped target can't
// block, and so owes nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "24c193cc-3f83-4414-8cac-8b9f48d5bb6d",
		Name:         "Grappling Hook",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{GrantToAttached("double strike")},
		Triggered: []game.TriggeredAbility{
			Optional(game.TriggeredAbility{
				Watches: []game.EventKind{game.EventAttack},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return attachedCreatureAttacked(ev, source)
				},
				Targets: TargetCreature("target creature"),
				Key:     "Grappling Hook — target creature blocks the equipped creature this turn if able",
				Effect:  TargetBlocksTheAttacker,
			}, "Grappling Hook — have target creature block the equipped creature this turn if able?"),
		},
		Activated: []ActivatedAbility{EquipAbility("{4}")},
	})
}
