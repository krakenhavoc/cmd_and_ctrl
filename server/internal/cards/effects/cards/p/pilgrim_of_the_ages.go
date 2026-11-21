package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pilgrim of the Ages — Creature — Spirit {2}{W}, 2/1 (Jumpstart 2022):
//
//	"When this creature enters, you may search your library for a basic
//	 Plains card, reveal it, put it into your hand, then shuffle.
//	 {6}: Return this card from your graveyard to your hand."
//
// Gloryheath Lynx's Plains search made optional (the prompt's decline skips
// the shuffle) and Clay Revenant's graveyard return at {6}.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f67084bb-07f8-420a-ad6e-9ab9842d3ee6",
		Name:         "Pilgrim of the Ages",
		Purpose:      game.Purpose{Tutors: 1},
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Pilgrim of the Ages — search for a basic Plains card", func(g *game.Game, item *game.StackItem) error {
				return effects.SearchLibrary{
					Player:    item.Controller,
					Predicate: effects.IsBasicLandWithSubtype("plains"),
					Dest:      game.ZoneHand,
					Limit:     1,
					Reveal:    true,
					Shuffle:   true,
					Optional:  true,
					Reason:    "Pilgrim of the Ages — a basic Plains card",
				}.Apply(effects.NewContext(g, item))
			}),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{6}: Return this card from your graveyard to your hand.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.ManaCost("{6}"),
			Zones:   []game.ZoneKind{game.ZoneGraveyard},
			Effect:  effects.ReturnThisCardFromYourGraveyardToYourHand,
		}},
	})
}
