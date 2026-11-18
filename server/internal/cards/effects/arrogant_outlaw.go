package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Arrogant Outlaw — Creature — Vampire Noble {2}{B}, 3/2 (Innistrad: Midnight
// Hunt, #5497):
//
//	"When this creature enters, if an opponent lost life this turn, each
//	 opponent loses 2 life and you gain 2 life."
//
// Bloodtithe Collector's intervening "if" (CR 603.4) with Brotherhood
// Patriarch's drain.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f202712d-7cd1-47aa-9664-526a884a1be2",
		Name:         "Arrogant Outlaw",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, AllOf(Self, func(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return opponentLifeLostThisTurn(g, source.Controller) > 0
			}), "Arrogant Outlaw — each opponent loses 2 life and you gain 2 life (an opponent lost life this turn)",
				Do(If{
					Cond: s82MidB06CondAnOpponentLostLife{},
					Then: []Applier{EachOpponentLosesLife{N: 2}, GainLife{Amount: 2}},
				})),
		},
	})
}
