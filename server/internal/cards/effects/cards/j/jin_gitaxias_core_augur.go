package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Jin-Gitaxias, Core Augur — Legendary Creature — Phyrexian Praetor
// {8}{U}{U}, 5/4:
//
//	"Flash
//	 At the beginning of your end step, draw seven cards.
//	 Each opponent's maximum hand size is reduced by seven."
//
// The reduction is a maximum-hand-size static (ADR 0113 §3, #2074),
// folded in CR 613.11's timestamp order with every other one that
// reaches the player. Alone it leaves each opponent a maximum of zero
// (the 2017-11-17 ruling), so they discard their whole hand in their
// own cleanup step (CR 514.1); it never touches its controller, who
// may still discard down to seven after the draw. Two of them make -7,
// which CR 107.1b turns into zero.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "eb23aed0-c450-4e57-96f2-2866dceca004",
		Name:            "Jin-Gitaxias, Core Augur",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			effects.AtYourEndStep("Jin-Gitaxias, Core Augur — draw seven cards", effects.Do(effects.DrawCards{N: 7})),
		},
		HandSize: []game.HandSizeStatic{effects.EachOpponentsMaxHandSizeReducedBy(7)},
	})
}
