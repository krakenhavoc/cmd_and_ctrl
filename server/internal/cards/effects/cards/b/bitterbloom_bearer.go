package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bitterbloom Bearer — Creature — Faerie Rogue {B}{B}, 1/1 (EDHREC
// rank 1988):
//
//	"Flash
//	 Flying
//	 At the beginning of your upkeep, you lose 1 life and create a
//	 1/1 blue and black Faerie creature token with flying."
//
// Bitterblossom on a body: the same upkeep trigger, life loss first
// and then the token. The token is a blue-and-black Faerie, not
// Bitterblossom's black Faerie Rogue, so it has its own template.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "5effb651-f9e0-4c14-8192-bd0e132b8d5d",
		Name:            "Bitterbloom Bearer",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash", "flying"},
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Bitterbloom Bearer — lose 1 life, create a Faerie", effects.Do(
				// The Faerie waits for a prompt the life loss asks (CR 608.2c, #4155).
				effects.Instruction(func(ctx *effects.Context) error {
					return ctx.Game.ChangePlayerLifeForEffect(ctx.Item.SourceCardID, ctx.Controller(), -1)
				}),
				effects.CreateToken{Template: effects.TokenCard("1/1 blue and black Faerie with flying"), N: 1},
			)),
		},
	})
}
