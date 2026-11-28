package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Barbed Sliver — Creature — Sliver {2}{R}, 2/2 (Tempest, S82 OLD b756, #5628):
//
//	"All Sliver creatures have "{2}: This creature gets +1/+0 until end of
//	 turn.""
//
// Firewake Sliver's granted activated ability (ADR 0093) on every Sliver
// creature at the table: the ability is the grantee's own, so "this creature"
// is the Sliver that has it, and only its controller can activate it.
//
// No simplification.
const barbedSliverGrant = "barbed-sliver/pump"

func init() {
	effects.Register(effects.Spec{
		OracleID:     "8fbca9ba-8b83-40ca-9f7e-5c483ba76675",
		Name:         "Barbed Sliver",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key: barbedSliverGrant,
			Activated: []effects.ActivatedAbility{{
				Label:   "{2}: This creature gets +1/+0 until end of turn",
				Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Power: 1}},
				Cost:    effects.ManaCost("{2}"),
				Effect:  effects.ThisGetsUntilEndOfTurn(1, 0, "Barbed Sliver — +1/+0 until end of turn"),
			}},
			Text: "{2}: This creature gets +1/+0 until end of turn.",
		}},
		Static: []game.StaticAbility{
			effects.TribalAbilityGrant(effects.TribeFilter{Tribes: []string{"Sliver"}}, barbedSliverGrant),
		},
	})
}
