package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kemba, Kha Regent — Legendary Creature — Cat Cleric {1}{W}{W}, 2/4
// (Scars of Mirrodin, #5577):
//
//	"At the beginning of your upkeep, create a 2/2 white Cat creature
//	 token for each Equipment attached to Kemba."
//
// Captain America, Liberator's count (msc16EquipmentAttachedToTheSource),
// read as the trigger resolves: an Equipment moved away in response no
// longer counts, and a Kemba that has left the battlefield counts what was
// attached to it as it last existed there (CR 608.2h; the card's ruling
// says the same). "Cat" and "white" are words of the card's text
// (ctx.TextToken).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "499d5801-0c56-40f2-9a46-77025ed6cb87",
		Name:         "Kemba, Kha Regent",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Kemba, Kha Regent — create a 2/2 white Cat for each Equipment attached to Kemba",
				effects.Do(effects.Instruction(func(ctx *effects.Context) error {
					n := effects.Msc16EquipmentAttachedToTheSource(ctx)
					if n <= 0 {
						return nil
					}
					return effects.CreateToken{Template: ctx.TextToken(effects.TokenCard("2/2 white Cat")), N: n}.Apply(ctx)
				}))),
		},
	})
}
