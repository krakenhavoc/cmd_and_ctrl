package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Squad Commander — Creature — Kor Warrior {3}{W}, 3/3 (Commander Legends:
// Battle for Baldur's Gate, #5948):
//
//	"When this creature enters, create a 1/1 white Kor Warrior creature token
//	 for each creature in your party. (Your party consists of up to one each
//	 of Cleric, Rogue, Warrior, and Wizard.)
//	 At the beginning of combat on your turn, if you have a full party,
//	 creatures you control get +1/+0 and gain indestructible until end of
//	 turn."
//
// The party is Shepherd of Heroes' (finB9PartySize, CR 700.8), counted as the
// trigger resolves, the Commander being the Warrior if nothing else is. The
// combat trigger is an intervening "if" (CR 603.4, IfYouHaveAFullParty): a
// full party is four creatures (CR 700.8c), read as combat begins and again
// as it resolves. Both effects affect the creatures you control as it
// resolves (CR 611.2c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "11f21b50-b4b6-449e-a66a-eedd2920ed93",
		Name:         "Squad Commander",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1, CreatureTokens: 1},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Squad Commander — create a 1/1 white Kor Warrior creature token for each creature in your party",
				effects.Do(effects.Instruction(func(ctx *effects.Context) error {
					return effects.CreateToken{Template: effects.TokenCard("1/1 white Kor Warrior"), N: effects.FinB9PartySize(ctx.Game, ctx.Controller())}.Apply(ctx)
				}))),
			effects.IfYouHaveAFullParty(effects.AtBeginningOfYourCombat("Squad Commander — creatures you control get +1/+0 and gain indestructible until end of turn",
				effects.Do(
					effects.BoostUntilEOT{Match: effects.And(effects.Creature(), effects.YouControl()), Power: 1, Label: "Squad Commander — +1/+0"},
					effects.GrantKeywordUntilEOT{Match: effects.And(effects.Creature(), effects.YouControl()), Keywords: []string{"indestructible"}, Label: "Squad Commander — indestructible"},
				))),
		},
	})
}
