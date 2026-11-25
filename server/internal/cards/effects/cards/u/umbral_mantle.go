package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Umbral Mantle — Artifact — Equipment {3}:
//
//	"Equipped creature has "{3}, {Q}: This creature gets +2/+2 until end of
//	 turn." ({Q} is the untap symbol.)
//	 Equip {0}"
//
// An ADR 0093 grant (Paradise Mantle's shape): the ability is the equipped
// creature's own, so moving the Mantle moves it. {Q} is AbilityCost.UntapSelf
// (CR 107.6), paid by the creature: it must be tapped, and a creature that has
// not been under its controller's control since their turn began cannot
// (CR 302.6). Each activation is its own +2/+2 that lasts until end of turn.
//
// No simplification.
const umbralMantleGrant = "umbral-mantle/plus-two-plus-two"

func init() {
	const label = "{3}, {Q}: This creature gets +2/+2 until end of turn."
	effects.Register(effects.Spec{
		OracleID:     "44f3241f-af82-4f2e-b15d-7f279bf6e816",
		Name:         "Umbral Mantle",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key: umbralMantleGrant,
			Activated: []effects.ActivatedAbility{{
				Label:   label,
				Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Power: 2, Toughness: 2}},
				Cost:    game.AbilityCost{Mana: "{3}", UntapSelf: true},
				Effect:  effects.ThisGetsUntilEndOfTurn(2, 2, "Umbral Mantle — +2/+2 until end of turn"),
			}},
			Text: label,
		}},
		Static:    []game.StaticAbility{effects.GrantAbilitiesToAttached(umbralMantleGrant)},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{0}")},
	})
}
