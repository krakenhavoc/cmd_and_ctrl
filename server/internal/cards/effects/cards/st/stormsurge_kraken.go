package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Stormsurge Kraken — Creature — Kraken {3}{U}{U}, 5/5 (Commander 2014,
// #5553):
//
//	"Hexproof
//	 Lieutenant — As long as you control your commander, this creature gets
//	 +2/+2 and has "Whenever this creature becomes blocked, you may draw two
//	 cards.""
//
// Demon of Wailing Agonies' lieutenant (c14LieutenantPump, c14WhileLieutenant):
// the +2/+2 and the granted trigger are on only while you control a
// commander you own. "Becomes blocked" is EventBecomesBlocked, once per
// blocked attacker however many creatures block it (Lim-Dûl's Paladin's
// trigger), and the draw is a real "you may" asked as it resolves.
//
// No simplification.
func init() {
	const label = "Stormsurge Kraken — Lieutenant: you may draw two cards"
	draw := effects.OptionalDraw(effects.On(game.EventBecomesBlocked, effects.Self, label, effects.Do(effects.DrawCards{N: 2})),
		"Stormsurge Kraken — draw two cards?", 2)
	effects.Register(effects.Spec{
		OracleID:        "72bdc7ba-2906-436d-8405-6a8a3335fdae",
		Name:            "Stormsurge Kraken",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"hexproof"},
		Static:          []game.StaticAbility{effects.C14LieutenantPump()},
		Triggered:       []game.TriggeredAbility{effects.C14WhileLieutenant(draw)},
	})
}
