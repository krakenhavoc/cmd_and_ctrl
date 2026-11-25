package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bellowing Elk — Creature — Elk {3}{G}, 4/2 (Modern Horizons):
//
//	"As long as you had another creature enter the battlefield under your
//	 control this turn, this creature has trample and indestructible."
//
// Two layer 6 keyword grants to the Elk itself, both held to the turn's entry
// tally (Bristlebane Outrider's reading): a creature that entered and has since
// left still counts, one that entered under an opponent's control does not, and
// when the Elk itself entered this turn it needs a second entry ("another").
//
// No simplification.
func init() {
	when := func(target *game.Card, g *game.Game, source *game.Card) bool {
		return effects.SelfCreature(target, g, source) && effects.BristlebaneOutriderAnotherCreatureEntered(g, source)
	}
	trample := effects.KeywordGrant(when, "trample")
	trample.Label = "As long as you had another creature enter the battlefield under your control this turn, this creature has trample."
	indestructible := effects.KeywordGrant(when, "indestructible")
	indestructible.Label = "As long as you had another creature enter the battlefield under your control this turn, this creature has indestructible."
	effects.Register(effects.Spec{
		OracleID:     "d17a1109-3a7f-4a39-b602-155da819235f",
		Name:         "Bellowing Elk",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{trample, indestructible},
	})
}
