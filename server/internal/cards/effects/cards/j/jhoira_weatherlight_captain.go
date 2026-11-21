package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Jhoira, Weatherlight Captain — Legendary Creature — Human Artificer
// {2}{U}{R}, 3/3 (EDHREC rank 1071):
//
//	"Whenever you cast a historic spell, draw a card. (Artifacts,
//	 legendaries, and Sagas are historic.)"
//
// The artifact-storm commander. "Historic" is CR 205.4h — an
// artifact, a legendary, or a Saga — read off the spell on the stack
// through b09IsHistoric. The trigger goes on the stack above the
// spell and resolves first, so the card is drawn before the historic
// spell lands, as in paper. Jhoira does not trigger on herself: she
// is on the stack, not the battlefield, when her own cast event
// fires.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c803b788-4213-4fab-b841-7e5bbf66088e",
		Name:         "Jhoira, Weatherlight Captain",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, effects.YouCastAHistoricSpell, "Jhoira, Weatherlight Captain — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
	})
}
