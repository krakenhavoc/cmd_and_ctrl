package effects

import (
	"encoding/json"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// mana_trigger_judge_test.go — ADR 0074's amendment of 2026-10-30
// (#4716), slice 1: the catalog half of the judge.
//
// The auto-tap planner is going to ask every triggered mana ability
// "would you fire for this production?" before anything is tapped, under
// a read lock, for a production that has not happened. An AppliesTo or a
// Produced that writes to the game, or that takes a lock a read lock
// cannot give it, would corrupt the plan or hang the table. The doc on
// game.ManaTrigger says they are read-only; this test is what makes the
// doc true for every card in the catalog today, and what fails the first
// card that is not.
//
// The delayed records a spell sets up (High Tide, Bubbling Muck, Molten
// Tide: game.DelayedTrigger.ManaTapSubtype) carry no callback at all, so
// there is nothing in them to break; the board still holds a few, so the
// judge's walk over them runs beside every card's.

// judgeProbe is one production the judge is asked about: a permanent
// tapped for the mana it names, whose pool is its controller's.
type judgeProbe struct {
	name   string
	source game.Card
	pool   uuid.UUID
	colors []string
}

// judgeBoard builds, for one catalogued key, a board the key's trigger
// has something to say about: the card itself on each of two seats (an
// Aura on that seat's Forest, which is also where a global trigger
// can sit), every land type and colour the family reads, a non-land
// source, two lands that share a name (Sasaya), and a few delayed
// records. `chosen` is the colour Caged Sun and Utopia Sprawl read;
// `monarch` makes seat 0 the monarch (Regal Behemoth).
func buildJudgeBoard(t *testing.T, key, chosen string, monarch bool) (*game.Game, []judgeProbe, []uuid.UUID) {
	t.Helper()
	g := newCatalogGame(t)
	seats := []*game.Player{g.Seats[0], g.Seats[1]}

	oracle, face := key, 0
	if i := strings.LastIndex(key, "#"); i >= 0 {
		n, err := strconv.Atoi(key[i+1:])
		if err != nil {
			t.Fatalf("catalog key %q: %v", key, err)
		}
		oracle, face = key[:i], n
	}

	var probes []judgeProbe
	var triggerSources []uuid.UUID
	for si, p := range seats {
		// Lands, as the tapped permanent was when it produced (a value).
		land := func(name, typeLine string) game.Card {
			c := game.Card{InstanceID: uuid.New(), Name: name, TypeLine: typeLine, Owner: p.ID, Controller: p.ID}
			pushBattlefieldCardWithTimestamp(g, c)
			return c
		}
		forest := land("Forest", "Basic Land — Forest")
		swamp := land("Swamp", "Basic Land — Swamp")
		land("Swamp", "Basic Land — Swamp") // a second Swamp: "each other land with the same name"
		island := land("Island", "Basic Land — Island")
		mountain := land("Mountain", "Basic Land — Mountain")
		plains := land("Plains", "Basic Land — Plains")
		snow := land("Snow-Covered Island", "Basic Snow Land — Island")
		wastes := land("Wastes", "Basic Land")
		dual := land("Overgrown Tomb", "Land — Swamp Forest")
		filter := land("Mystic Gate", "Land")
		ring := game.Card{InstanceID: uuid.New(), Name: "Sol Ring", TypeLine: "Artifact", Owner: p.ID, Controller: p.ID}
		pushBattlefieldCardWithTimestamp(g, ring)

		probes = append(probes,
			judgeProbe{"Forest {G}", forest, p.ID, []string{"G"}},
			judgeProbe{"Swamp {B}", swamp, p.ID, []string{"B"}},
			judgeProbe{"Island {U}", island, p.ID, []string{"U"}},
			judgeProbe{"Mountain {R}", mountain, p.ID, []string{"R"}},
			judgeProbe{"Plains {W}", plains, p.ID, []string{"W"}},
			judgeProbe{"snow Island {U}", snow, p.ID, []string{"U"}},
			judgeProbe{"Wastes {C}", wastes, p.ID, []string{"C"}},
			judgeProbe{"Wastes {C}{C}", wastes, p.ID, []string{"C", "C"}},
			judgeProbe{"Swamp Forest {B}", dual, p.ID, []string{"B"}},
			judgeProbe{"Swamp Forest {G}", dual, p.ID, []string{"G"}},
			judgeProbe{"filter land {U}{R}", filter, p.ID, []string{"U", "R"}},
			judgeProbe{"Sol Ring {C}{C}", ring, p.ID, []string{"C", "C"}},
		)

		// The card under test, on this seat: an Aura on this seat's Forest.
		id := uuid.New()
		c := game.Card{
			InstanceID: id, Name: "Trigger " + key, TypeLine: "Enchantment — Aura", OracleID: oracle,
			ActiveFace: face, Owner: p.ID, Controller: p.ID, ChosenColor: chosen,
			AttachedTo: game.TargetRef{Kind: game.TargetCard, ID: forest.InstanceID},
		}
		pushBattlefieldCardWithTimestamp(g, c)
		triggerSources = append(triggerSources, id)
		if si == 0 && monarch {
			g.WithWriteLock(func() { g.Monarch = p.ID })
		}
	}

	// Delayed records, as a resolved High Tide / Bubbling Muck / Molten
	// Tide leave them: plain data the judge walks first.
	g.WithWriteLock(func() {
		for _, dt := range []game.DelayedTrigger{
			{Controller: seats[0].ID, SourceCardID: uuid.New(), Label: "turn trigger", ManaTapSubtype: "Island", ManaAdds: "{U}", Repeats: true},
			{Controller: seats[1].ID, SourceCardID: uuid.New(), Label: "your turn trigger", ManaTapSubtype: "Swamp", ManaAdds: "{B}", Repeats: true, ManaTapControllerOnly: true},
		} {
			g.ScheduleDelayedTriggerForEffect(dt)
		}
	})
	return g, probes, triggerSources
}

// persistedBytes is the game's persistence format as bytes, minus the
// one field that is the wall clock. The event log, the pools, the
// prompts, every permanent and the delayed queue are in it.
func persistedBytes(t *testing.T, g *game.Game) []byte {
	t.Helper()
	s := g.CaptureSnapshot()
	s.TakenAt = time.Time{}
	blob, err := json.Marshal(s)
	if err != nil {
		t.Fatalf("marshal snapshot: %v", err)
	}
	return blob
}

// askTheJudge is the judge under a read lock, as the planner will ask
// it. It runs on its own goroutine so that a callback waiting for a lock
// a read lock cannot give it fails the test rather than hanging it.
func askTheJudge(t *testing.T, g *game.Game, prod game.ManaProduced, where string) []uuid.UUID {
	t.Helper()
	done := make(chan []uuid.UUID, 1)
	go func() {
		var out []uuid.UUID
		g.ReadSnapshot(func() { out = g.ManaTriggerSourcesForTest(prod) })
		done <- out
	}()
	select {
	case out := <-done:
		return out
	case <-time.After(20 * time.Second):
		t.Fatalf("%s: the judge did not return under a read lock: a callback is taking a lock", where)
		return nil
	}
}

// TestEveryCatalogManaTriggerIsReadOnly: for every catalogued card with
// a triggered mana ability, on a board built for it, asking the judge
// under ReadSnapshot about a spread of productions leaves the game's
// snapshot byte-identical, does not edit the production it was handed,
// and returns (it does not wait on a lock a read lock cannot give).
func TestEveryCatalogManaTriggerIsReadOnly(t *testing.T) {
	if len(productionDefKeys) == 0 {
		t.Fatal("TestMain did not capture the production catalog keys")
	}
	var keys []string
	triggers := 0
	for _, key := range productionDefKeys {
		if d := defs[key]; d != nil && len(d.ManaTriggers) > 0 {
			keys = append(keys, key)
			triggers += len(d.ManaTriggers)
		}
	}
	// A floor, not a count: the catalog grows. It catches a lookup that
	// silently stopped finding the cards (an empty loop passes).
	if len(keys) < 15 {
		t.Fatalf("only %d catalogued cards have a triggered mana ability, want at least 15: has the lookup broken?", len(keys))
	}

	type setting struct {
		chosen  string
		monarch bool
	}
	settings := []setting{{"B", true}, {"", false}}
	judged, firings := 0, 0

	for _, key := range keys {
		fired := 0
		for _, set := range settings {
			where := fmt.Sprintf("%s (chosen %q, monarch %v)", key, set.chosen, set.monarch)
			g, probes, ownSources := buildJudgeBoard(t, key, set.chosen, set.monarch)

			// The board must actually carry the key's triggers, or every
			// check below passes on a card the engine does not see.
			var seen int
			g.ReadSnapshot(func() {
				for _, c := range g.BattlefieldCardsForEffect() {
					if slices.Contains(ownSources, c.InstanceID) {
						seen += len(game.ManaTriggersForCard(c))
					}
				}
			})
			if want := len(defs[key].ManaTriggers) * len(ownSources); seen != want {
				t.Fatalf("%s: the board shows %d mana triggers on the card, want %d: the probe card does not resolve to its catalog key",
					where, seen, want)
			}

			// Settle the layers once, so ReadSnapshot has no recompute to do
			// between the two snapshots.
			g.ReadSnapshot(func() {})
			before := persistedBytes(t, g)
			for _, pr := range probes {
				prod := game.ManaProduced{Source: pr.source, Controller: pr.pool, Colors: slices.Clone(pr.colors)}
				got := askTheJudge(t, g, prod, where+", "+pr.name)
				if !slices.Equal(prod.Colors, pr.colors) || prod.Source.InstanceID != pr.source.InstanceID || prod.Controller != pr.pool {
					t.Errorf("%s, %s: a callback edited the production it was handed: %+v", where, pr.name, prod)
				}
				judged++
				for _, source := range got {
					if slices.Contains(ownSources, source) {
						fired++
						firings++
					}
				}
			}
			if after := persistedBytes(t, g); string(before) != string(after) {
				// Name the production: replay on a fresh board, one at a time.
				g2, probes2, _ := buildJudgeBoard(t, key, set.chosen, set.monarch)
				g2.ReadSnapshot(func() {})
				offender := "(none alone: the changes add up)"
				for _, pr := range probes2 {
					was := persistedBytes(t, g2)
					askTheJudge(t, g2, game.ManaProduced{Source: pr.source, Controller: pr.pool, Colors: slices.Clone(pr.colors)}, where)
					if string(was) != string(persistedBytes(t, g2)) {
						offender = pr.name
						break
					}
				}
				t.Errorf("%s: the judge changed the game's snapshot, first on %s", where, offender)
			}
		}
		// Not vacuous: the matrix has to make the card fire somewhere, or
		// its Produced was never run and a write in it was never seen.
		if fired == 0 {
			t.Errorf("%s: no production in the matrix made its trigger fire, so its Produced callback was never exercised; extend buildJudgeBoard", key)
		}
	}
	t.Logf("%d catalogued cards, %d triggered mana abilities, %d judge calls, %d firings of their own", len(keys), triggers, judged, firings)
}
