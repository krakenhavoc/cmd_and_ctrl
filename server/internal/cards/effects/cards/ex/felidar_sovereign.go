package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Felidar Sovereign — Creature — Cat Beast {4}{W}{W}, 4/6:
//
//	"Vigilance
//	 Lifelink
//	 At the beginning of your upkeep, if you have 40 or more life, you
//	 win the game."
//
// ADR 0057's reference effect win (#749, CR 104.2b). The upkeep
// trigger is an intervening "if" (CR 603.4): it triggers only if its
// controller has 40 or more life as the upkeep begins, and it checks
// again as it resolves — so a Lightning Bolt in response that drops
// them to 37 leaves the trigger doing nothing, which is the printed
// outcome.
//
// In Commander the starting life total is 40, so the Sovereign wins in
// its controller's next upkeep unless somebody has done something about
// it. That is the card, not a bug.
//
// The win is immediate, during the resolution, and it ends the game for
// everyone (CR 104.1): the other players are not eliminated, the game
// is simply over, and GameView.outcome names the winner and the
// Sovereign. An opponent's Platinum Angel or Herald of Eternal Dawn —
// or the controller's own Abyssal Persecutor — prevents it; the log
// says so and the game goes on.
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "5ae3cbb9-9f0c-4077-ae7c-fba660d7fb4b",
		Name:            "Felidar Sovereign",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance", "lifelink"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBeginUpkeep, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.Actor == source.Controller && effects.LifeAtLeast(g, source.Controller, 40)
			}, "Felidar Sovereign — if you have 40 or more life, you win the game", effects.WinIfLifeAtLeast(40)),
		},
	})
}
