package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dogged Detective — Creature — Human Rogue {1}{B}, 2/1:
//
//	"When this creature enters, surveil 2.
//	 Whenever an opponent draws their second card each turn, you may return
//	 this card from your graveyard to your hand."
//
// The second ability watches from the graveyard only (InGraveyard, CR 113.6),
// where "you" and "an opponent" are the card's owner's. The draw tally is the
// drawing player's, read as the draw is recorded, so the second draw of the
// turn reports exactly two and every later one more
// (b41DrewYourSecondCardThisTurn's rule, for an opponent). It is a "may"
// asked as the trigger resolves, and does nothing if the card has left the
// graveyard by then.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c5a1516a-5e01-48c5-88c7-2178f4b2ad24",
		Name:         "Dogged Detective",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Dogged Detective — surveil 2", effects.Do(effects.Surveil{N: 2})),
			effects.InGraveyard(effects.Optional(effects.On(game.EventDrawCard,
				func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
					return effects.ByAnOpponent(ev, source, lki, g) && g.TurnTallyFor(ev.Actor).CardsDrawn == 2
				},
				"Dogged Detective — you may return it from your graveyard to your hand",
				effects.ReturnThisCardFromYourGraveyardToHand),
				"Return Dogged Detective from your graveyard to your hand?")),
		},
	})
}
