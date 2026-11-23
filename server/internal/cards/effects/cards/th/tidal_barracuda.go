package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tidal Barracuda — Creature — Fish {3}{U}, 3/4 (Commander 2020, #5513):
//
//	"Any player may cast spells as though they had flash.
//	 Your opponents can't cast spells during your turn."
//
// The first line is Vedalken Orrery's CastAsThoughFlash declared for every
// player (game.TimingAffectsEachPlayer), the Barracuda's controller
// included; the second is Dragonlord Dromoka's restriction. A restriction
// beats a permission (CR 101.2), so on your turn your opponents cast nothing
// and on theirs everyone may cast at instant speed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "edfc0a46-414e-467d-9f27-e42d08ec127a",
		Name:         "Tidal Barracuda",
		Completeness: effects.CompletenessFull,
		CastTimings: []game.CastTimingRule{{
			Timing:  game.TimingFlash,
			Affects: game.TimingAffectsEachPlayer,
			Label:   "Any player may cast spells as though they had flash.",
		}},
		CastRestrictions: []game.CastRestriction{
			effects.OpponentsCantCastDuringYourTurn("Tidal Barracuda — your opponents can't cast spells during your turn.", nil),
		},
	})
}
