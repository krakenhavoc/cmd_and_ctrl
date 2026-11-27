package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Archery Training — {W} Enchantment — Aura (Urza's Destiny, S82 OLD batch 703, #5621):
//
//	"Enchant creature
//	 At the beginning of your upkeep, you may put an arrow counter on this
//	 Aura.
//	 Enchanted creature has "{T}: This creature deals X damage to target
//	 attacking or blocking creature, where X is the number of arrow counters
//	 on Archery Training.""
//
// A granted activated ability (Rime Transfusion's shape). The creature is the
// source of the damage; X is read from the Aura that granted the ability
// (StackItem.GrantedBy), so two Archery Trainings each count their own.
//
// No simplification.
const archeryText = "{T}: This creature deals X damage to target attacking or blocking creature, where X is the number of arrow counters on Archery Training."

func init() {
	effects.Register(effects.Spec{
		OracleID:     "acab4fa9-d9ab-4a89-a50f-0cee9d7ca0c3",
		Name:         "Archery Training",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Grants: []effects.AbilityGrant{{
			Key: "archery-training/shoot",
			Activated: []effects.ActivatedAbility{{
				Label:   archeryText,
				Purpose: game.Purpose{Answers: game.AnswerRemove},
				Cost:    effects.TapCost(),
				Targets: effects.TargetCreature("target attacking or blocking creature", effects.AttackingOrBlocking()),
				Effect:  effects.S82OldB703ArcheryTrainingShot,
			}},
			Text: archeryText,
		}},
		Static: []game.StaticAbility{effects.GrantAbilitiesToAttached("archery-training/shoot")},
		Triggered: []game.TriggeredAbility{
			effects.S82OldB703MayPutCounterOnThisAtUpkeep("Archery Training", "arrow"),
		},
	})
}
