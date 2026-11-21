package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Arcbound Crusher — Artifact Creature — Juggernaut {4}, 0/0 (Darksteel,
// #5605):
//
//	"Trample
//	 Whenever another artifact enters, put a +1/+1 counter on this
//	 creature.
//	 Modular 1 (This creature enters with a +1/+1 counter on it. When it
//	 dies, you may put its +1/+1 counters on target artifact creature.)"
//
// "Another artifact" is any artifact under anyone's control, Grinding
// Station's watcher (b20ArtifactEntered) less the Crusher's own entry.
// The counter goes on only if the Crusher is still on the battlefield as
// the trigger resolves. Trample and modular are engine keywords (modular:
// #2012, CR 702.43a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ee496b34-d8e0-482a-a2c2-3122ea376da0",
		Name:            "Arcbound Crusher",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample", "modular 1"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, anotherArtifactEntered, "Arcbound Crusher — put a +1/+1 counter on this creature",
				effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}

// anotherArtifactEntered is "whenever another artifact enters": an
// artifact other than the source, under any player's control.
func anotherArtifactEntered(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	return ev.CardID != source.InstanceID && effects.B20ArtifactEntered(ev, g)
}
