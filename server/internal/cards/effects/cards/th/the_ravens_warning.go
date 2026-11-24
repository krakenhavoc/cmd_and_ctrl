package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Raven's Warning — Enchantment — Saga {1}{W}{U} (Kaldheim, #5505):
//
//	"I — Create a 1/1 blue Bird creature token with flying. You gain 2 life.
//	 II — Whenever one or more creatures you control with flying deal combat
//	 damage to a player this turn, look at that player's hand and draw a
//	 card.
//	 III — You may put a card you own from outside the game on top of your
//	 library."
//
// Chapter II is a repeating delayed trigger (CR 603.7b) over an
// each-condition (CR 603.2c) that names the player the combat damage
// reached, fired once per player in a combat damage step however many
// flyers connected. It outlives the Saga for the rest of the turn; the look
// is private to the Saga's controller.
//
// Chapter III is a card you own from outside the game (CR 400.11, ADR 0191):
// the owner's sideboard, or the card database for a deck that seated none
// (ADR 0198). Declining, or an empty pool, does nothing. The card goes on
// top without a shuffle and without being revealed.
//
// No simplification.
func init() {
	const name = "The Raven's Warning"
	effects.Register(effects.Spec{
		OracleID:     "74206de7-470b-4256-9207-9fb46d3c4356",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1, CreatureTokens: 1},
		Triggered: []game.TriggeredAbility{
			effects.ChapterTrigger(1, effects.SagaChapterLabel(name, 1, "create a 1/1 blue Bird creature token with flying; you gain 2 life"),
				effects.Do(
					effects.CreateToken{Template: effects.TokenCard("1/1 blue Bird with flying"), N: 1},
					effects.GainLife{Amount: 2},
				)),
			effects.ChapterTrigger(2, effects.SagaChapterLabel(name, 2, "whenever one or more creatures you control with flying deal combat damage to a player this turn, look at that player's hand and draw a card"),
				func(g *game.Game, item *game.StackItem) error {
					return effects.DelayedOnEvent{
						Label:        name + " — look at that player's hand and draw a card",
						On:           []game.EventKind{game.EventDealDamage},
						Condition:    effects.OldB059FlyingCreatureHitAPlayerCondition,
						Body:         effects.OldB059LookAtThatHandAndDrawBody,
						Repeats:      true,
						OncePerBatch: true,
					}.Apply(effects.NewContext(g, item))
				}),
			effects.ChapterTrigger(3, effects.SagaChapterLabel(name, 3, "you may put a card you own from outside the game on top of your library"),
				func(g *game.Game, item *game.StackItem) error {
					g.QueueOutsidePickForEffect(game.OutsidePickPrompt{
						Player:   item.Controller,
						Source:   item.SourceCardID,
						Question: name + " — you may put a card you own from outside the game on top of your library",
						Min:      0,
						Max:      1,
						Keyed:    theRavensWarningTop,
					})
					return nil
				}),
		},
	})
}

// theRavensWarningTop is chapter III's pick: the one card (or none) goes on
// top of its owner's library. An on-disk key.
var theRavensWarningTop = game.RegisterOutsidePick("choose-cards/the-ravens-warning-top", nil,
	func(g *game.Game, r game.CardsChosen) error {
		if !r.Answered {
			return nil
		}
		picked := g.StillOutsideForEffect(r.Chooser, r.Picked)
		if len(picked) == 0 {
			return nil
		}
		_, err := g.PutFromOutsideOnLibraryTopForEffect(r.Chooser, picked[0])
		return err
	})
