package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Manriki-Gusari — Artifact — Equipment {2} (Saviors of Kamigawa, S82 OLD b593):
//
//	"Equipped creature gets +1/+2 and has "{T}: Destroy target Equipment."
//	 Equip {1}"
//
// A granted activated ability (ADR 0093): the equipped creature's controller
// activates it, and it needs no tap-symbol exemption beyond the creature's own
// summoning sickness (CR 302.6).
//
// No simplification.
const manrikiGusariGrant = "manriki-gusari/destroy-target-equipment"

func init() {
	effects.Register(effects.Spec{
		OracleID:     "cc2eb3e4-084c-471b-9fe0-72aa9ff2019a",
		Name:         "Manriki-Gusari",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key: manrikiGusariGrant,
			Activated: []effects.ActivatedAbility{{
				Label:   "{T}: Destroy target Equipment",
				Purpose: game.Purpose{Answers: game.AnswerRemove},
				Cost:    effects.TapCost(),
				Targets: effects.TargetPermanent("target Equipment", effects.HasSubtype("Equipment")),
				Effect:  effects.DestroyFirstLegalCardTarget,
			}},
			Text: "{T}: Destroy target Equipment.",
		}},
		Static:    []game.StaticAbility{effects.PumpAttached(1, 2), effects.GrantAbilitiesToAttached(manrikiGusariGrant)},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{1}")},
	})
}
