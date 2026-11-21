package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// undying_persist_test.go — #2075, CR 702.93 and CR 702.79 (ADR 0113
// §4). Undying and persist are canonical keyword tokens whose dies
// triggers the engine derives from the departed permanent's last-known
// ability list (game/undying_persist.go). One test per interaction the
// ADR lists, then one per card that lands with the seam.

const (
	OracleGleefulArsonist      = "9025b3ce-f6ae-431d-9eb5-02d713b8c9b8"
	OraclePersistentConstrictr = "bf0ee9da-6071-440c-9149-d441f730fcfb"
	OracleMurderousRedcap      = "a498bc70-36e7-4454-bc44-906893df38b8"

	// undyingLordOracle is a test-only lord: "Other creatures you
	// control have undying." Mikaeus, the Unhallowed's grant, without
	// the rest of Mikaeus.
	UndyingLordOracle = "test-2075-undying-lord"
)

func init() {
	effects.Register(effects.Spec{
		OracleID: UndyingLordOracle,
		Name:     "Undying Lord",
		Static: []game.StaticAbility{
			effects.TribalKeywordGrant(effects.TribeFilter{Others: true, YoursOnly: true}, game.KeywordUndying),
		},
	})
}

// pushDiesKeywordCreature is a creature with no catalog entry whose
// printed keywords the deck importer stamped — Young Wolf's and Safehold
// Elite's case, the proof that they need no card file.
func PushDiesKeywordCreature(g *game.Game, controller uuid.UUID, name string, power, toughness int, keywords ...string) uuid.UUID {
	return PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: "Creature — Wolf",
		Power: power, Toughness: toughness, Keywords: keywords,
		Owner: controller, Controller: controller,
	})
}

// diesReturnItems counts the undying and persist triggers waiting on
// the stack or headed for it.
func DiesReturnItems(g *game.Game) int {
	n := 0
	is := func(item *game.StackItem) bool {
		return item != nil && (item.Body == "undying/return" || item.Body == "persist/return")
	}
	for _, item := range g.StackMeta {
		if is(item) {
			n++
		}
	}
	for _, item := range g.PendingTriggers {
		if is(item) {
			n++
		}
	}
	return n
}

func AddCounters(t *testing.T, g *game.Game, id uuid.UUID, kind string, n int) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.AddCounterForEffect(id, kind, n); err != nil {
			t.Fatalf("AddCounterForEffect: %v", err)
		}
	})
}

func EffectErrors(g *game.Game) []string {
	var out []string
	g.ReadSnapshot(func() {
		for _, ev := range g.Events {
			if ev.Kind == game.EventEffectError {
				out = append(out, ev.ErrorMsg)
			}
		}
	})
	return out
}
