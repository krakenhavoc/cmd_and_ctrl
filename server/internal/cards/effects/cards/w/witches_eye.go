package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

const witchesEyeGrant = "witches-eye/scry"

// Witches' Eye — Artifact — Equipment {1} (Theros, S82 old sets):
//
//	"Equipped creature has "{1}, {T}: Scry 1."
//	 Equip {1}"
//
// An ADR 0093 grant (Viridian Longbow's shape): the scry is the equipped
// creature's own ability, so its {T} cost keeps CR 302.6's summoning-sickness
// rule, and moving the Eye moves the ability.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c8d479b8-b96c-4382-97d3-68d85cf371ba",
		Name:         "Witches' Eye",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key: witchesEyeGrant,
			Activated: []effects.ActivatedAbility{{
				Label:   "{1}, {T}: Scry 1.",
				Purpose: game.Purpose{Answers: game.AnswerValue, Scry: 1},
				Cost:    effects.Plus(effects.ManaCost("{1}"), effects.TapCost()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.Scry{Player: item.Controller, N: 1}.Apply(effects.NewContext(g, item))
				},
			}},
			Text: "{1}, {T}: Scry 1.",
		}},
		Static:    []game.StaticAbility{effects.GrantAbilitiesToAttached(witchesEyeGrant)},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{1}")},
	})
}
