package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Keral Keep Disciples — Creature — Human Monk {2}{R}{R}, 4/3 (Core Set
// 2021, #5511):
//
//	"Whenever you activate a loyalty ability of a Chandra planeswalker, this
//	 creature deals 1 damage to each opponent."
//
// The trigger reads the activation event's loyalty bit (CR 606.2) and the
// activated planeswalker's Chandra subtype through the layered view, so a
// planeswalker that is a Chandra by an effect counts and a card merely named
// Chandra does not. "You" is the one who activated it. One activation is one
// trigger; the damage is the Disciples' own, dealt to each opponent at once.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0a865453-ddc8-4403-8f1e-5752e3582458",
		Name:         "Keral Keep Disciples",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventActivateAbility, keralKeepDisciplesActivated,
				"Keral Keep Disciples — it deals 1 damage to each opponent",
				effects.Do(effects.DamageEachOpponent{N: 1})),
		},
	})
}

// keralKeepDisciplesActivated is "you activate a loyalty ability of a Chandra
// planeswalker".
func keralKeepDisciplesActivated(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventActivateAbility || !ev.Loyalty || ev.Actor != source.Controller {
		return false
	}
	pw, ok := g.LookupCardForEffect(ev.CardID)
	return ok && pw.IsPlaneswalker() && pw.HasSubtype("Chandra")
}
