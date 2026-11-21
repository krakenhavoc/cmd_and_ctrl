package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bishop of Wings — Creature — Human Cleric {W}{W}, 1/4 (EDHREC rank
// 3824):
//
//	"Whenever an Angel you control enters, you gain 4 life.
//	 Whenever an Angel you control dies, create a 1/1 white Spirit
//	 creature token with flying."
//
// The Angel deck's glue. Two triggers, both on Angels the controller
// controls — effective subtypes, so a changeling counts. The Bishop
// is a Human Cleric and never matches either; the dies trigger reads
// the Angel as it last existed on the battlefield (CR 603.10a,
// #1679), so a creature that was an Angel only through a grant
// counts too.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "af4684cf-f109-44be-adfb-7c551a36635e",
		Name:         "Bishop of Wings",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.B21AngelYouControlEntered(ev, source, g)
			}, "Bishop of Wings — you gain 4 life", effects.Do(effects.GainLife{Amount: 4})),
			effects.On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.B36AngelYouControlDied(ev, source, g)
			}, "Bishop of Wings — create a 1/1 white Spirit with flying", effects.Do(effects.CreateToken{Template: effects.B28WhiteSpiritFlyingToken(), N: 1})),
		},
	})
}
