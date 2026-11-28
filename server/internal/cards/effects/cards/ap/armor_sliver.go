package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Armor Sliver — Creature — Sliver {2}{W}, 2/2 (Tempest, S82 OLD b756, #5628):
//
//	"All Sliver creatures have "{2}: This creature gets +0/+1 until end of
//	 turn.""
//
// Barbed Sliver's granted pump with the toughness bonus.
//
// No simplification.
const armorSliverGrant = "armor-sliver/pump"

func init() {
	effects.Register(effects.Spec{
		OracleID:     "c284b1fd-f238-478a-bcf1-c4b8f32f11f3",
		Name:         "Armor Sliver",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key: armorSliverGrant,
			Activated: []effects.ActivatedAbility{{
				Label:   "{2}: This creature gets +0/+1 until end of turn",
				Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Toughness: 1}},
				Cost:    effects.ManaCost("{2}"),
				Effect:  effects.ThisGetsUntilEndOfTurn(0, 1, "Armor Sliver — +0/+1 until end of turn"),
			}},
			Text: "{2}: This creature gets +0/+1 until end of turn.",
		}},
		Static: []game.StaticAbility{
			effects.TribalAbilityGrant(effects.TribeFilter{Tribes: []string{"Sliver"}}, armorSliverGrant),
		},
	})
}
