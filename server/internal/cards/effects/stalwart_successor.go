package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Stalwart Successor — Creature — Human Warrior {1}{B}{G}, 3/2 (Tarkir:
// Dragonstorm, #4463):
//
//	"Menace
//	 Whenever one or more counters are put on a creature you control, if
//	 it's the first time counters have been put on that creature this turn,
//	 put a +1/+1 counter on that creature."
//
// The seam is Game.TimesCountersPutThisTurn (turn_tally.go): per OBJECT,
// how many times counters have been put on a permanent this turn, written
// by the tally listener ahead of the trigger harvester, so a placement
// that reads 1 is the first (Captain America's tap counter, for counters).
// A counter of any kind counts, a removal does not, counters a creature
// enters with count (CR 122.6), and a permanent that left and came back
// is a new object (CR 400.7). The Successor's own +1/+1 counter is the
// second placement, so it triggers nothing more. The "if" is a fact about
// the placement that triggered, decided when it happens. The counter goes
// on the creature the event named if it is still there.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "45392882-e4b9-4c74-a10f-985c2b4e2be9",
		Name:            "Stalwart Successor",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			On(game.EventCounterPlaced, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.Target == uuid.Nil || ev.Amount <= g.CounterCountBeforeForEffect(ev) {
					return false
				}
				if z := g.FindCardZoneForEffect(ev.Target); z == nil || z.Kind != game.ZoneBattlefield {
					return false
				}
				c, ok := g.LookupCardForEffect(ev.Target)
				return ok && c.IsCreature() && c.Controller == source.Controller &&
					g.TimesCountersPutThisTurn(ev.Target) == 1
			}, "Stalwart Successor — put a +1/+1 counter on that creature",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					target := ctx.Trigger().Event.Target
					if !onBattlefield(g, target) {
						return nil
					}
					return AddCounter{Target: target, Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
				}),
		},
	})
}
