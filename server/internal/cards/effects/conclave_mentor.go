package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Conclave Mentor — Creature — Centaur Cleric {G}{W}, 2/2 (EDHREC
// rank 2344):
//
//	"If one or more +1/+1 counters would be put on a creature you
//	 control, that many plus one +1/+1 counters are put on that
//	 creature instead.
//	 When this creature dies, you gain life equal to its power."
//
// Hardened Scales on a body. The replacement is the Scales' exactly:
// +1/+1 counters only, creatures the controller controls only, the
// Mentor itself included, CounterDelta bumped by one. The dies
// trigger reads the Mentor's last-known power — the harvester's LKI
// characteristic carries the layer-computed P/T, and the +1/+1
// counters it grew (its own effect makes those common) are read back
// off the log, b13LastKnownPower's job.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a2fe5937-212c-4e71-8d6e-f408b38100aa",
		Name:         "Conclave Mentor",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{{
			Watches: []game.EventKind{game.EventCounterPlaced},
			AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
				if ev.Kind != game.RepEventCounter || ev.CounterName != "+1/+1" || ev.CounterDelta <= 0 {
					return false
				}
				target, ok := g.LookupCardForEffect(ev.CounterTarget)
				return ok && target.IsCreature() && target.Controller == src.Controller
			},
			Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
				ev.CounterDelta++
				return nil
			},
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				return src.Controller
			},
			Label: "Conclave Mentor: +1 +1/+1 counter",
		}},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventLTB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return cardDied(ev, source)
			},
			Build: func(_ game.Event, source *game.Card, lki game.Characteristic, g *game.Game) *game.StackItem {
				power := b13LastKnownPower(g, source.InstanceID, lki)
				return game.NewTriggeredItem(source, "Conclave Mentor — gain life equal to its power",
					func(g *game.Game, item *game.StackItem) error {
						return GainLife{Player: item.Controller, Amount: power}.Apply(NewContext(g, item))
					})
			},
		}},
	})
}
