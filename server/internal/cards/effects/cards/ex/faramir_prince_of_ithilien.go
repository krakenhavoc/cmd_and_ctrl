package cardsex

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const faramirLabel = "Faramir, Prince of Ithilien — choose an opponent; at the beginning of that player's next end step, you draw a card if they didn't attack you that turn, otherwise create three 1/1 white Human Soldier creature tokens"

// Faramir, Prince of Ithilien — Legendary Creature — Human Noble {2}{W}{U},
// 3/3 (The Lord of the Rings: Tales of Middle-earth Commander):
//
//	"At the beginning of your end step, choose an opponent. At the
//	 beginning of that player's next end step, you draw a card if they
//	 didn't attack you that turn. Otherwise, create three 1/1 white Human
//	 Soldier creature tokens."
//
// The choice is a keyed choose-a-player (Rowan Kenrith's keyed delayed
// trigger on a named player's turn): the delayed ability waits for the
// chosen player's next end step (CR 603.7), stays controlled by Faramir's
// controller (CR 603.7d) and reads, as it resolves, whether that player
// attacked its controller that turn (an attack on a planeswalker or battle
// is not an attack on the player). Faramir leaving does not stop it, and
// Faramir's controller gets the card or the tokens.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8a8c7c20-beb7-4352-99ea-7f64a30b9f52",
		Name:         "Faramir, Prince of Ithilien",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 3, CreatureTokens: 3, Draws: 1},
		Triggered: []game.TriggeredAbility{
			effects.AtYourEndStep(faramirLabel, effects.Do(effects.ChoosePlayer{
				Among:    effects.Opponents,
				Question: "Faramir, Prince of Ithilien — choose an opponent",
				ThenKey:  faramirChosen,
			})),
		},
	})
}

var faramirChosen = effects.ChoosePlayerThen("option-pick/faramir-prince-of-ithilien-opponent", func(ctx *effects.Context, chosen uuid.UUID, _ []uuid.UUID) error {
	if chosen == uuid.Nil {
		return nil
	}
	return effects.ScheduleDelayedTrigger{
		At:     game.StepEnd,
		TurnOf: chosen,
		Label:  "Faramir, Prince of Ithilien — draw a card if that player didn't attack you this turn, otherwise create three 1/1 white Human Soldier tokens",
		Body:   faramirEndStepBody,
		Params: game.EffectParams{Player: chosen},
	}.Apply(ctx)
})

var faramirEndStepBody = game.DelayedBody("faramir-prince-of-ithilien/end-step",
	func(g *game.Game, item *game.StackItem, p game.EffectParams) error {
		ctx := effects.NewContext(g, item)
		if effects.GapB127PlayerAttackedYouThisTurn(g, p.Player, item.Controller) {
			return effects.CreateToken{Controller: item.Controller, Template: effects.TokenCard("1/1 white Human Soldier"), N: 3}.Apply(ctx)
		}
		return effects.DrawCards{Player: item.Controller, N: 1}.Apply(ctx)
	})
