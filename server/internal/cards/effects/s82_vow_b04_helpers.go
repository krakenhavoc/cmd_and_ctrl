package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_vow_b04_helpers.go — shared pieces of the Crimson Vow Blood-token
// cards (S82, #5495): reading a Blood token sacrifice off the event, and
// counting Blood tokens.

// vowB04YouSacrificedBlood is "Whenever you sacrifice a Blood token": the
// sacrifice's actor is the source's controller and the permanent that
// went (read last-known, so a Blood paid as its own cost still counts,
// CR 701.21) is a Blood.
func VowB04YouSacrificedBlood(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventSacrifice || ev.Actor != source.Controller {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && c.HasSubtypeWord("Blood")
}

// vowB04BloodTokensControlledBy counts the Blood tokens `player` controls.
func VowB04BloodTokensControlledBy(g *game.Game, player uuid.UUID) int {
	return CountControlled(g, player, func(c game.Card) bool {
		return c.IsToken() && c.HasSubtypeWord("Blood")
	})
}

// vowB04BloodAtEndStepIfLifeGained is the end-step trigger both faces of
// Restless Bloodseeker print: "At the beginning of your end step, if you
// gained life this turn, create a Blood token." The intervening "if"
// (CR 603.4) is checked as the trigger fires and again as it resolves.
func VowB04BloodAtEndStepIfLifeGained(name string) game.TriggeredAbility {
	return On(game.EventBeginEndStep, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
		return B24YourEndStepAndYouGainedLifeThisTurn(ev, source, g)
	}, name+" — create a Blood token (you gained life this turn)",
		Do(If{
			Cond: CondLifeGained{},
			Then: []Applier{CreateToken{Template: BloodToken(), N: 1}},
		}))
}
