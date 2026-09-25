package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bristly Bill, Spine Sower — Legendary Creature — Plant Druid,
// {1}{G}, 2/2 (EDHREC rank 921):
//
//	"Landfall — Whenever a land you control enters, put a +1/+1
//	 counter on target creature.
//	 {3}{G}{G}: Double the number of +1/+1 counters on each creature
//	 you control."
//
// The two-drop counters commander. Landfall is a targeted trigger —
// "target creature", any creature at the table, Bill included — so
// the controller picks when it goes on the stack and the pick is
// re-checked at resolution. The activated ability doubles what is
// already there: for each creature you control it adds as many +1/+1
// counters as it has, from a snapshot taken before the first is
// placed, so a Doubling Season on the first creature cannot change
// what the second receives (and does itself double each placement,
// as printed).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d3b2d8a2-d3bc-448c-9cf6-6bead6010c28",
		Name:         "Bristly Bill, Spine Sower",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := enteredUnderYourControl(ev, source, g, false)
				return ok && c.IsLand()
			},
			Targets: TargetCreature("target creature"),
			Key:     "Bristly Bill, Spine Sower — a +1/+1 counter on target creature",
			Effect: func(g *game.Game, item *game.StackItem) error {
				if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
					return nil
				}
				return AddCounter{Target: item.Targets[0].ID, Kind: "+1/+1", N: 1}.Apply(NewContext(g, item))
			},
		}},
		Activated: []ActivatedAbility{{
			Label:  "{3}{G}{G}: Double the number of +1/+1 counters on each creature you control.",
			Cost:   ManaCost("{3}{G}{G}"),
			Effect: b08DoubleCountersOnEachCreatureYouControl,
		}},
	})
}
