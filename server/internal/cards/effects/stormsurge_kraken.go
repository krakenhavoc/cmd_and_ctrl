package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	draw := OptionalDraw(On(game.EventBecomesBlocked, Self, label, Do(DrawCards{N: 2})),
		"Stormsurge Kraken — draw two cards?", 2)
	Register(Spec{
		OracleID:        "72bdc7ba-2906-436d-8405-6a8a3335fdae",
		Name:            "Stormsurge Kraken",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"hexproof"},
		Static:          []game.StaticAbility{c14LieutenantPump()},
		Triggered:       []game.TriggeredAbility{c14WhileLieutenant(draw)},
	})
}
