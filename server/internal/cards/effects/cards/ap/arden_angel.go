package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Arden Angel — Creature — Angel {4}{W}{W}, 4/4 (Secret Lair Drop):
//
//	"Flying
//	 At the beginning of your upkeep, if Arden Angel is in your graveyard, roll
//	 a four-sided die. If the result is 1, return Arden Angel from your
//	 graveyard to the battlefield."
//
// A graveyard upkeep trigger (Squee's shape): the intervening "if" is the
// trigger being the card's graveyard trigger, and the card is checked to be
// the same object still in the graveyard as the roll resolves. The d4 goes
// through the engine's one dice door; the result is the natural roll.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "82d51c8a-d42b-48b3-a726-3a8d28d057e5",
		Name:            "Arden Angel",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.InGraveyard(effects.AtYourUpkeep("Arden Angel — roll a four-sided die; on a 1, return it from your graveyard to the battlefield",
				effects.Do(effects.RollDice{Sides: 4, N: 1, Then: ardenAngelRolled}))),
		},
	})
}

var ardenAngelRolled = effects.DiceTail("arden-angel/rolled", func(ctx *effects.Context, rolls []int) error {
	if rolls[0] != 1 || !effects.CardObjectStillInGraveyard(ctx.Game, ctx.Item) {
		return nil
	}
	return effects.ReturnThisCardFromYourGraveyard(ctx.Game, ctx.Item)
})
