package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Chance Encounter — Enchantment {2}{R}{R}:
//
//	"Whenever you win a coin flip, put a luck counter on this enchantment.
//	 At the beginning of your upkeep, if this enchantment has ten or more
//	 luck counters on it, you win the game."
//
// One luck counter per coin won (WheneverYouWinACoinFlip, #3767). The
// upkeep trigger is an intervening if (CR 603.4), read as the upkeep
// begins and again as it resolves, so removing counters in response
// stops the win. The win is ADR 0057's effect win, which a "can't win
// the game" gate prevents.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8bd33b5c-0777-41ef-9e17-fb30ce69ec25",
		Name:         "Chance Encounter",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverYouWinACoinFlip("Chance Encounter — put a luck counter on it", func(g *game.Game, item *game.StackItem) error {
				return AddCounter{Target: item.SourceCardID, Kind: "luck", N: 1}.Apply(NewContext(g, item))
			}),
			On(game.EventBeginUpkeep, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Actor == source.Controller && source.Counters["luck"] >= 10
			}, "Chance Encounter — if it has ten or more luck counters, you win the game", chanceEncounterWin),
		},
	})
}

// chanceEncounterWin re-checks the intervening if on the same object
// (CR 603.4, CR 400.7) and wins the game.
func chanceEncounterWin(g *game.Game, item *game.StackItem) error {
	if item.Controller == uuid.Nil || sourceIsNewObject(g, item) || !onBattlefield(g, item.SourceCardID) {
		return nil
	}
	c, ok := g.LookupCardForEffect(item.SourceCardID)
	if !ok || c.Counters["luck"] < 10 {
		return nil
	}
	return WinTheGame{Player: item.Controller}.Apply(NewContext(g, item))
}
