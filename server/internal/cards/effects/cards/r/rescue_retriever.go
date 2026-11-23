package cardsr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rescue Retriever — Creature — Dog Soldier {3}{W}{W}, 3/3 (The Brothers'
// War, #5482):
//
//	"Flash
//	 When this creature enters, put a +1/+1 counter on each other Soldier
//	 you control.
//	 Prevent all damage that would be dealt to other attacking Soldiers you
//	 control."
//
// The counters go on the Soldiers on the battlefield as the trigger
// resolves, each its own placement, never on the Retriever. The last line
// is a standing CR 615 prevention over every damage event, combat or not,
// aimed at an attacking Soldier you control other than the Retriever.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "86f9897f-84aa-4c8d-8335-cfdb2f565311",
		Name:            "Rescue Retriever",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Rescue Retriever — a +1/+1 counter on each other Soldier you control",
				rescueRetrieverCounters),
		},
		Replacements: []game.ReplacementEffect{
			effects.BroB05PreventDamageToOtherAttackingSoldiers("Rescue Retriever: prevent all damage that would be dealt to other attacking Soldiers you control"),
		},
	})
}

// rescueRetrieverCounters puts a +1/+1 counter on each Soldier the
// ability's controller controls other than its source.
func rescueRetrieverCounters(g *game.Game, item *game.StackItem) error {
	source := item.SourceCardID
	reader, ok := g.LookupCardForEffect(source)
	if !ok {
		return nil
	}
	// "Soldier" is a word of the Retriever's text (CR 612.2).
	return effects.PutPlusOneCountersOnEachYouControlThat(1,
		func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
			return c.InstanceID != source && game.ThroughTextOf(&reader, c).HasSubtypeWord("Soldier")
		})(g, item)
}
