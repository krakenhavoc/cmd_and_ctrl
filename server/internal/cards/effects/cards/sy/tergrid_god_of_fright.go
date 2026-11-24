package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tergrid, God of Fright // Tergrid's Lantern — modal double-faced card,
// {3}{B}{B} // {3}{B} (Kaldheim).
//
// Front, Legendary Creature — God, 4/5:
//
//	"Menace
//	 Whenever an opponent sacrifices a nontoken permanent or discards a
//	 permanent card, you may put that card from a graveyard onto the
//	 battlefield under your control."
//
// One triggered ability watching two events (a sacrifice and a discard); the
// card is the event's own card, put onto the battlefield from whichever
// graveyard it is in if it still is (CR 608.2b), under the Tergrid's
// controller's control. "You may" is asked as the trigger resolves.
//
// Back, Legendary Artifact (registered under "<oracle_id>#1", ADR 0034):
//
//	"{T}: Target player loses 3 life unless they sacrifice a nonland
//	 permanent of their choice or discard a card.
//	 {3}{B}: Untap Tergrid's Lantern."
//
// The target player picks among the loss and each alternative they can do.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "8485cfaa-1dbf-432b-b5d0-92a6aa6a329b",
		Name:            "Tergrid, God of Fright",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.OnAny([]game.EventKind{game.EventSacrifice, game.EventDiscardCard},
				effects.S82OldB062OpponentSacrificedNontokenOrDiscardedPermanent,
				"Tergrid, God of Fright — put that card from a graveyard onto the battlefield under your control",
				effects.S82OldB062PutThatCardOntoTheBattlefield),
				"Tergrid, God of Fright — put that card onto the battlefield under your control?"),
		},
	})
	effects.Register(effects.Spec{
		OracleID:     "8485cfaa-1dbf-432b-b5d0-92a6aa6a329b#1",
		Name:         "Tergrid's Lantern",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{T}: Target player loses 3 life unless they sacrifice a nonland permanent of their choice or discard a card.",
				Cost:    effects.TapCost(),
				Targets: effects.TargetPlayer("target player"),
				Effect:  effects.S82OldB062LanternEffect,
			},
			{
				UntapSelf: true,
				Label:     "{3}{B}: Untap Tergrid's Lantern.",
				Purpose:   game.Purpose{Answers: game.AnswerValue},
				Cost:      effects.ManaCost("{3}{B}"),
				Effect:    effects.UntapThisPermanentEffect,
			},
		},
	})
}
