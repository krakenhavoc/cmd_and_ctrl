package cardshu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Infiltration Lens — Artifact — Equipment {1} (Scars of Mirrodin, #5577):
//
//	"Whenever equipped creature becomes blocked by a creature, you may
//	 draw two cards.
//	 Equip {1}"
//
// "Becomes blocked by a creature" triggers once per blocker (CR 509.3d),
// so it watches the per-pair EventBlock, as Dead-Iron Sledge does, with
// the equipped creature as the attacker (Target) rather than the blocker.
// Two blockers are two triggers and four cards. The draw is optional.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "72dfd729-ebba-4bb1-83eb-54c2051fea3b",
		Name:         "Infiltration Lens",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.OptionalDraw(effects.On(game.EventBlock, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Target != uuid.Nil && source.IsAttachedTo(ev.Target)
			}, "Infiltration Lens — you may draw two cards", effects.Do(effects.DrawCards{N: 2})),
				"Infiltration Lens — draw two cards?", 2),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{1}")},
	})
}
