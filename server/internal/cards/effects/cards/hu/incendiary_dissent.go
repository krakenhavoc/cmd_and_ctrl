package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Incendiary Dissent — Conspiracy (Conspiracy: Take the Crown, S82 OLD batch
// 268, #5540):
//
//	"Hidden agenda
//	 Creatures you control with the chosen name have "{R}: This creature gets
//	 +1/+0 until end of turn.""
//
// Hidden agenda (ADR 0227). A layer-6 grant of a catalog bundle (ADR 0093) to
// the creatures with the chosen name; the granted activated ability is the
// creature's own, so "this creature" is the recipient and its controller pays.
//
// No simplification.
func init() {
	const grantKey = "s82-old-b268/incendiary-dissent-firebreathing"
	effects.Register(effects.Spec{
		OracleID:     "70caaa64-9623-4d05-9843-18460fea963b",
		Name:         "Incendiary Dissent",
		Completeness: effects.CompletenessFull,
		HiddenAgenda: effects.HiddenAgendaOneName,
		Grants: []effects.AbilityGrant{{
			Key: grantKey,
			Activated: []effects.ActivatedAbility{{
				Label:   "{R}: This creature gets +1/+0 until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Power: 1}},
				Cost:    effects.ManaCost("{R}"),
				Effect:  effects.ThisGetsUntilEndOfTurn(1, 0, "Incendiary Dissent — +1/+0 until end of turn"),
			}},
			Text: "{R}: This creature gets +1/+0 until end of turn.",
		}},
		Static: []game.StaticAbility{effects.GrantAbilities(effects.CreaturesYouControlWithTheChosenName, grantKey)},
	})
}
