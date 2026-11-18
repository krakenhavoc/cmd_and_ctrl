package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cloaked Cadet — Creature — Human Ranger {4}{G}, 2/4 (Crimson Vow):
//
//	"Training (Whenever this creature attacks with another creature with
//	 greater power, put a +1/+1 counter on this creature.)
//	 Whenever one or more +1/+1 counters are put on one or more Humans you
//	 control, draw a card. This ability triggers only once each turn."
//
// The passive "are put on" counts every placement on a Human its controller
// controls, whoever puts it (Shalai and Hallar's reading), the Cadet's own
// training counter included (it is a Human). "Only once each turn" is the
// per-turn trigger record Ant-Man, Colony Commander and Exemplar of Light read.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "08a40717-c272-42e6-adb1-0148614b1202",
		Name:         "Cloaked Cadet",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Training("Cloaked Cadet — training: put a +1/+1 counter on it"),
			On(game.EventCounterPlaced, cloakedCadetCounterPlaced, cloakedCadetDrawLabel, Do(DrawCards{N: 1})),
		},
	})
}

const cloakedCadetDrawLabel = "Cloaked Cadet — draw a card"

func cloakedCadetCounterPlaced(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if source == nil {
		return false
	}
	if _, ok := b39PlusOneCountersOnCreatureYouControl(ev, source, g); !ok {
		return false
	}
	if target, ok := g.LookupCardForEffect(ev.Target); !ok || !target.HasSubtypeWord("Human") {
		return false
	}
	return !b11TriggeredThisTurn(g, source.InstanceID, cloakedCadetDrawLabel)
}
