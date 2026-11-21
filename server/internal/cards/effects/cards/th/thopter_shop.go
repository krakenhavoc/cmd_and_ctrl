package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Thopter Shop — Artifact {3} (The Brothers' War Commander):
//
//	"Whenever one or more artifact creatures you control die, draw a card.
//	 This ability triggers only once each turn.
//	 {2}{W}, {T}: Create a 1/1 colorless Thopter artifact creature token
//	 with flying."
//
// The first line is one batch trigger per death event (CR 603.2c), reading
// the dying permanent's last-known types and controller (CR 603.10a), and
// limited to once each turn. The second is a plain token maker.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b2f9e07b-64b7-40b3-9a5e-5f1d59b35af7",
		Name:         "Thopter Shop",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TriggersOnlyOnceEachTurn(effects.OncePerBatch(effects.On(game.EventLTB, thopterShopArtifactCreatureYouControlDied,
				"Thopter Shop — draw a card",
				effects.Do(effects.DrawCards{N: 1})))),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{W}, {T}: Create a 1/1 colorless Thopter artifact creature token with flying.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Tokens: 1, CreatureTokens: 1},
			Cost:    effects.Plus(effects.ManaCost("{2}{W}"), effects.TapCost()),
			Effect:  effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 colorless Thopter artifact with flying"), N: 1}),
		}},
	})
}

// thopterShopArtifactCreatureYouControlDied is one death in the batch: an
// artifact creature Thopter Shop's controller controlled as it died.
func thopterShopArtifactCreatureYouControlDied(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	dead, ok := effects.DiedCreature(ev, g)
	return ok && effects.LeftAsType(ev, dead, "artifact") && effects.LeftUnderControlOf(ev, dead) == source.Controller
}
