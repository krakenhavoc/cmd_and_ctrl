package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vash, Freelance Archaeologist — Legendary Creature — Human Rogue
// Scientist {1}{B}{G}, 2/4 (Star Trek, #4973):
//
//	"At the beginning of your upkeep, surveil 1.
//	 Whenever a card leaves your graveyard, draw a card. This ability
//	 triggers only once each turn."
//
// The upkeep trigger is the plain surveil (CR 701.25a). The second
// trigger watches every way a card leaves its controller's graveyard,
// a cast from it included (leftGraveyardEvents, cardLeftYourGraveyard:
// the graveyard is the one Vash's controller owns, and a token is not a
// card). "Only once each turn" is Curator of Sun's Creation's check: the
// harvester counts the trigger the moment it is queued, so a second card
// leaving later in the turn, or alongside the first in the same event,
// does not trigger it again (OncePerBatch keeps one event's several
// cards to one trigger).
//
// No simplification.
func init() {
	const draw = "Vash, Freelance Archaeologist — a card left your graveyard: draw a card"
	effects.Register(effects.Spec{
		OracleID:     "ec5a9dde-b89d-4c54-8563-4a7170f5cefd",
		Name:         "Vash, Freelance Archaeologist",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Vash, Freelance Archaeologist — surveil 1", effects.Do(effects.Surveil{N: 1})),
			effects.OncePerBatch(effects.OnAny(effects.LeftGraveyardEvents, effects.AllOf(
				func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return effects.CardLeftYourGraveyard(ev, source, g)
				}, effects.CuratorFirstThisTurn(draw)), draw, effects.Do(effects.DrawCards{N: 1}))),
		},
	})
}
