package effects

import (
	"github.com/google/uuid"

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
	Register(Spec{
		OracleID:        "5ae3cbb9-9f0c-4077-ae7c-fba660d7fb4b",
		Name:            "Felidar Sovereign",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance", "lifelink"},
		Triggered: []game.TriggeredAbility{
			On(game.EventBeginUpkeep, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.Actor == source.Controller && lifeAtLeast(g, source.Controller, 40)
			}, "Felidar Sovereign — if you have 40 or more life, you win the game", winIfLifeAtLeast(40)),
		},
	})
}

// lifeAtLeast reports whether `player` is still in the game with at
// least `n` life — the intervening-if of the "at the beginning of your
// upkeep, if you have N or more life, you win the game" family (Felidar
// Sovereign's 40, Test of Endurance's 50).
func lifeAtLeast(g *game.Game, player uuid.UUID, n int) bool {
	p := g.PlayerByIDForEffect(player)
	return p != nil && !p.Eliminated && p.Life >= n
}

// winIfLifeAtLeast is that family's effect: the intervening-if checked
// again on resolution (CR 603.4), then WinTheGame. Captures only the
// threshold, so it survives Clone and undo.
func winIfLifeAtLeast(n int) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		if !lifeAtLeast(g, item.Controller, n) {
			return nil
		}
		return WinTheGame{Player: item.Controller}.Apply(NewContext(g, item))
	}
}
