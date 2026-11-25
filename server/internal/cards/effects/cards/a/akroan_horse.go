package cardsa

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Akroan Horse — Artifact Creature — Horse {4}, 0/4 (Commander 2016, S82 older
// sets, #5537):
//
//	"Defender
//	 When this creature enters, an opponent gains control of it.
//	 At the beginning of your upkeep, each opponent creates a 1/1 white
//	 Soldier creature token."
//
// Not a target: its controller picks the opponent as the trigger resolves
// (CR 608.2d; with one opponent there is nothing to pick), the control change
// has no duration (CR 611.2a), and "your upkeep" and "each opponent" are then
// read for the new controller, so the Soldiers go to the player who used to
// control it and to every other opponent of the new controller.
//
// No simplification.
func init() {
	akroanHorseOpponentChosen = effects.OptionPickThen("option-pick/akroan-horse-opponent", akroanHorseGainControl)
	effects.Register(effects.Spec{
		OracleID:        "8d0213d7-9e25-49d2-93fa-1d346cb2e869",
		Name:            "Akroan Horse",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Akroan Horse — an opponent gains control of it", akroanHorseOffer),
			effects.AtYourUpkeep("Akroan Horse — each opponent creates a 1/1 white Soldier creature token",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					for _, opp := range effects.OpponentsInTurnOrder(g, item.Controller) {
						if err := (effects.CreateToken{Controller: opp, Template: effects.TokenCard("1/1 white Soldier"), N: 1}).Apply(ctx); err != nil {
							return err
						}
					}
					return nil
				}),
		},
	})
}

// akroanHorseOpponentChosen is the keyed answer to "which opponent". Assigned
// in init: the key is an on-disk identity, never renamed, never reused.
var akroanHorseOpponentChosen game.OptionPickThen

func akroanHorseOffer(g *game.Game, item *game.StackItem) error {
	if !effects.SourceIsStillThisPermanent(g, item) {
		return nil
	}
	opponents := effects.OpponentsInTurnOrder(g, item.Controller)
	switch len(opponents) {
	case 0:
		return nil
	case 1:
		akroanHorseTakeControl(g, item.SourceCardID, opponents[0])
		return nil
	}
	g.QueueChoosePlayerForEffect(game.ChoosePlayerPrompt{
		Chooser:  item.Controller,
		Among:    opponents,
		Source:   item.SourceCardID,
		Question: "Akroan Horse — choose an opponent to gain control of it",
		ThenKey:  akroanHorseOpponentChosen,
	})
	return nil
}

// akroanHorseGainControl hands the Horse to the chosen opponent. Nobody
// chosen (the chooser left the game) changes nothing.
func akroanHorseGainControl(ctx *effects.Context, r game.OptionPicked) error {
	if r.Option == nil || r.Option.Player == uuid.Nil {
		return nil
	}
	akroanHorseTakeControl(ctx.Game, r.Source, r.Option.Player)
	return nil
}

func akroanHorseTakeControl(g *game.Game, horse, to uuid.UUID) {
	g.GainControlForEffect(horse, horse, to, game.IndefiniteDuration(), "Akroan Horse — an opponent gains control of it")
}
