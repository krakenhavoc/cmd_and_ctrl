package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// counter_each_test.go — #1841, CR 603.2c: "whenever a [kind] counter
// is put on ~" triggers once PER COUNTER, each its own stack object
// and its own "you may".

const (
	OracleFathomMage         = "93d0e129-e3b5-4aff-9e50-f34771ed00ff"
	OracleBloodcrazedHoplite = "b1f441ce-7619-4ab5-ab36-724ed0a76728"
	OracleFlourishingDef     = "f1b26eb1-6337-42d6-a475-53ad6427f00f"
	OraclePerCounterEntrant  = "00001841-0000-4000-8000-000000000001"
)

func init() {
	// A test-only card that ENTERS with two +1/+1 counters and has
	// Fathom Mage's ability, for the CR 122.6 entry case: no real
	// Fathom Mage enters with counters on its own.
	effects.Register(effects.Spec{
		OracleID: OraclePerCounterEntrant,
		Name:     "Per-Counter Entrant (test)",
		Replacements: []game.ReplacementEffect{{
			Watches: []game.EventKind{game.EventZoneMove},
			AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, src *game.Card) bool {
				return ev.Kind == game.RepEventMove && ev.NewZone == game.ZoneBattlefield &&
					src != nil && ev.CardID == src.InstanceID
			},
			Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
				ev.AddCounterAtETB(game.CounterPlusOne, 2)
				return nil
			},
			Label: "test: enters with two +1/+1 counters",
		}},
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.WheneverACounterIsPutOnThis(game.CounterPlusOne, "test — you may draw a card",
				func(g *game.Game, item *game.StackItem) error {
					return effects.DrawCards{Player: item.Controller, N: 1}.Apply(effects.NewContext(g, item))
				}), "test — draw a card?"),
		},
	})
}

func PutCounters(t *testing.T, g *game.Game, id uuid.UUID, kind string, n int) {
	t.Helper()
	if err := g.AddCounter(id, kind, n); err != nil {
		t.Fatalf("AddCounter: %v", err)
	}
}

// triggerPrompts is how many "you may"s `who` owes (owedMays).
func TriggerPrompts(g *game.Game, who uuid.UUID) int {
	return OwedMays(g, who)
}

// answerAllPrompts says yes to every waiting prompt, and returns how
// many there were.
func AnswerAllPrompts(t *testing.T, g *game.Game, who uuid.UUID) int {
	t.Helper()
	n := TriggerPrompts(g, who)
	for i := 0; i < n; i++ {
		AnswerLatestTriggerPrompt(t, g, who, true)
	}
	return n
}

func CeSeedLibrary(g *game.Game, id uuid.UUID, n int) {
	for i := 0; i < n; i++ {
		PushLibraryCardForTest(PlayerByIDForTest(g, id), game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Land", Owner: id})
	}
}
