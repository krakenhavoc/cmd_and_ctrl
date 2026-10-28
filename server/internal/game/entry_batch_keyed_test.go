package game

import (
	"bytes"
	"encoding/json"
	"errors"
	"reflect"
	"sort"
	"strconv"
	"testing"

	"github.com/google/uuid"
)

// entry_batch_keyed_test.go — #4404: a simultaneous entry of created
// tokens, and an Attraction's open, paused on one card's CR 616 order
// prompt, are restore points. Every restore is strict (RestoreStrict
// through JSON), taken while the prompt is open, and every restored answer
// is checked against the same answer given to a table that was never
// restored.

// twoEntersTapped stubs two Kismet-shaped entry replacements under
// `controller`: two effects on one entry under anyone else, so its
// controller orders them (CR 616.1).
func twoEntersTapped(t *testing.T, g *Game, controller uuid.UUID) {
	t.Helper()
	stubCatalogReplacements(t, map[string][]ReplacementEffect{
		"kismet": {entersTappedForOthers("Kismet: enters tapped")},
		"thalia": {entersTappedForOthers("Thalia: enters tapped")},
	})
	pushReplacementSource(g, "kismet", controller)
	pushReplacementSource(g, "thalia", controller)
}

// answerOnlyOrder answers the one open order prompt in the order offered.
func answerOnlyOrder(t *testing.T, g *Game, chooser uuid.UUID) {
	t.Helper()
	c := onlyReplacementOrderPrompt(t, g, chooser)
	if err := g.ResolveReplacementOrder(c.ID, chooser, c.ReplacementEffectIDs); err != nil {
		t.Fatalf("ResolveReplacementOrder: %v", err)
	}
}

// entryOutcome is what an entry left on the board, in terms two tables
// built the same way share: each permanent named `name`, tapped or not,
// with its counters; the events of `kinds`; and the open prompts.
type entryOutcome struct {
	Permanents []string
	Events     int
	Open       int
	Ran        int
}

func observeEntries(g *Game, name string, kind EventKind, ran int) entryOutcome {
	out := entryOutcome{Ran: ran}
	for _, c := range g.Battlefield.Cards {
		if c.Name == name {
			out.Permanents = append(out.Permanents, c.Name+" tapped="+strconv.FormatBool(c.Tapped)+" +1="+strconv.Itoa(c.Counters[CounterPlusOne]))
		}
	}
	sort.Strings(out.Permanents)
	for _, ev := range g.Events {
		if ev.Kind == kind {
			out.Events++
		}
	}
	for _, c := range g.PendingChoices {
		if c != nil {
			out.Open++
		}
	}
	return out
}

// tokensTogether creates two Goblins together under seat 1, each with a
// +1/+1 counter, while seat 0 controls two enters-tapped effects: the
// batch pauses on the first Goblin's order prompt. *ran counts the
// creation's continuation.
func tokensTogether(t *testing.T, ran *int) (*Game, uuid.UUID) {
	t.Helper()
	g := newActiveGame(t)
	me, them := g.Seats[0].ID, g.Seats[1].ID
	twoEntersTapped(t, g, me)
	g.WithWriteLock(func() {
		spec := TokenCreation{Controller: them, Together: true,
			Groups: []TokenGroup{{Template: goblinToken(), Count: 2, Entry: TokenEntryOptions{Counters: map[string]int{CounterPlusOne: 1}}}},
		}
		if err := g.CreateTokensKeyedForEffect(spec, hookTail(func(_ *Game, o TailOutcome) error {
			*ran += len(o.IDs)
			return nil
		})); err != nil {
			t.Fatal(err)
		}
	})
	onlyReplacementOrderPrompt(t, g, them)
	return g, them
}

// Created tokens entering together, paused on the first token's prompt and
// again on the second's, are restore points under the token-batch key, and
// the restored answers land both tokens together with their entry clause
// and run the creation's continuation once, as on a table never restored.
func TestTokensEnteringTogetherAreARestorePoint(t *testing.T) {
	liveRan := 0
	live, them := tokensTogether(t, &liveRan)
	answerOnlyOrder(t, live, them)
	answerOnlyOrder(t, live, them)
	want := observeEntries(live, "Goblin", EventTokenCreated, liveRan)
	if len(want.Permanents) != 2 || want.Ran != 2 || want.Permanents[0] != "Goblin tapped=true +1=1" {
		t.Fatalf("never restored: %+v; want two tapped Goblins with a counter, the continuation told both", want)
	}

	ran := 0
	g, them := tokensTogether(t, &ran)
	g = combatRestore(t, g, ReplacementResumeMoveTokenBatch)
	answerOnlyOrder(t, g, them)
	if n := tokensNamed(g, "Goblin"); n != 0 {
		t.Fatalf("%d Goblins landed before the second token's window settled", n)
	}
	g = combatRestore(t, g, ReplacementResumeMoveTokenBatch)
	answerOnlyOrder(t, g, them)
	if got := observeEntries(g, "Goblin", EventTokenCreated, ran); !reflect.DeepEqual(got, want) {
		t.Fatalf("restored %+v, never restored %+v", got, want)
	}
	if len(g.enteringTokens) != 0 {
		t.Errorf("%d tokens still staged after the batch landed", len(g.enteringTokens))
	}
}

// attractionOpen opens two Attractions for seat 0 while seat 1 controls
// two enters-tapped effects: the first open pauses on its order prompt.
func attractionOpen(t *testing.T) (*Game, uuid.UUID) {
	t.Helper()
	g := newAttractionGame(t, 1, attrDeck(10)...)
	me, opp := g.Seats[0].ID, g.Seats[1].ID
	twoEntersTapped(t, g, opp)
	g.WithWriteLock(func() {
		if err := g.OpenAttractionsForEffect(me, uuid.Nil, 2); err != nil {
			t.Fatal(err)
		}
	})
	onlyReplacementOrderPrompt(t, g, me)
	if g.attractionEntering == uuid.Nil {
		t.Fatal("no Attraction is entering")
	}
	return g, me
}

// attractionsOnBattlefield is each Attraction on the battlefield, tapped
// or not.
func attractionsOnBattlefield(g *Game) []string {
	var out []string
	for _, c := range g.Battlefield.Cards {
		if IsAttraction(c) {
			out = append(out, c.Name+" tapped="+strconv.FormatBool(c.Tapped))
		}
	}
	sort.Strings(out)
	return out
}

// An "open two Attractions" paused on the first open's prompt, and again
// on the second's, is a restore point under the Attraction key. The
// restored answers open both, tapped, each announcing its open, as on a
// table never restored.
func TestAnAttractionsOpenIsARestorePoint(t *testing.T) {
	live, me := attractionOpen(t)
	answerOnlyOrder(t, live, me)
	answerOnlyOrder(t, live, me)
	want := attractionsOnBattlefield(live)
	wantOpened := len(attrEvents(live, 0, EventAttractionOpened))
	if len(want) != 2 || !bytes.HasSuffix([]byte(want[0]), []byte("tapped=true")) || !bytes.HasSuffix([]byte(want[1]), []byte("tapped=true")) || wantOpened != 2 {
		t.Fatalf("never restored: %v, %d opened; want two tapped Attractions opened", want, wantOpened)
	}

	g, me := attractionOpen(t)
	g = combatRestore(t, g, ReplacementResumeMoveAttractionOpen)
	answerOnlyOrder(t, g, me)
	g = combatRestore(t, g, ReplacementResumeMoveAttractionOpen)
	answerOnlyOrder(t, g, me)
	if got := attractionsOnBattlefield(g); !reflect.DeepEqual(got, want) {
		t.Fatalf("restored %v, never restored %v", got, want)
	}
	if n := len(attrEvents(g, 0, EventAttractionOpened)); n != wantOpened {
		t.Errorf("restored table announced %d opens, want %d", n, wantOpened)
	}
	if len(g.Seats[0].AttractionDeckCards()) != len(live.Seats[0].AttractionDeckCards()) {
		t.Errorf("Attraction deck holds %d, want %d", len(g.Seats[0].AttractionDeckCards()), len(live.Seats[0].AttractionDeckCards()))
	}
	if g.attractionEntering != uuid.Nil {
		t.Error("the open's index window was left open")
	}
}

// The batch keys are refused out of place: a token batch or an open under
// the plain move-tail key, which a binary from before #4404 accepts.
func TestEntryBatchKeysAreRefusedOutOfPlace(t *testing.T) {
	ran := 0
	tokens, _ := tokensTogether(t, &ran)
	open, _ := attractionOpen(t)
	for name, tc := range map[string]struct {
		g   *Game
		key string
	}{
		"a token batch": {tokens, ReplacementResumeMoveTokenBatch},
		"an open":       {open, ReplacementResumeMoveAttractionOpen},
	} {
		raw, err := json.Marshal(tc.g.CaptureSnapshot())
		if err != nil {
			t.Fatal(err)
		}
		key := []byte(`"` + tc.key + `"`)
		if !bytes.Contains(raw, key) {
			t.Fatalf("%s: no %s key in the file", name, tc.key)
		}
		var s GameSnapshot
		if err := json.Unmarshal(bytes.Replace(raw, key, []byte(`"`+ReplacementResumeMoveTail+`"`), 1), &s); err != nil {
			t.Fatal(err)
		}
		if _, err := s.RestoreStrict(); !errors.Is(err, ErrUnknownEffectKey) {
			t.Errorf("%s under the move-tail key: restore = %v, want ErrUnknownEffectKey", name, err)
		}
	}
}

// The key each batch takes, and a frozen entry over either stays counted.
func TestAnEntryBatchOfItsOwnKindTakesItsOwnKey(t *testing.T) {
	g := NewGame()
	frame := func(b *entryBatch) *replacementResumeFrame {
		return &replacementResumeFrame{
			ev:         &ReplacementEvent{Kind: RepEventMove, CardID: uuid.New(), OldZone: ZoneCommand, NewZone: ZoneBattlefield, entryTail: &entryTail{batch: b}},
			applicable: []activeReplacement{{id: builtinReplacementIDBase}},
		}
	}
	for name, tc := range map[string]struct {
		b    *entryBatch
		want string
	}{
		"an ordinary batch": {&entryBatch{}, ReplacementResumeMoveTail},
		"a token batch":     {&entryBatch{members: []entryBatchMember{{cardID: uuid.New(), token: true}}}, ReplacementResumeMoveTokenBatch},
		"an open":           {&entryBatch{attraction: &attractionOpenRun{N: 2}}, ReplacementResumeMoveAttractionOpen},
	} {
		f := frame(tc.b)
		if key := g.replacementFrameKeyable(&PendingChoice{Kind: PendingChoiceReplacementOrder, replacementResume: f}); key != tc.want {
			t.Errorf("%s: keyed %q, want %q", name, key, tc.want)
		}
		if frozen := g.frozenEntryKeyable(f); frozen != (tc.want == ReplacementResumeMoveTail) {
			t.Errorf("%s: frozen entry keyable = %v", name, frozen)
		}
	}
}

// A token member's membership round-trips through the batch mirror.
func TestATokenBatchMemberRoundTrips(t *testing.T) {
	g := NewGame()
	b := &entryBatch{
		source: uuid.New(), next: 1, attraction: &attractionOpenRun{Player: uuid.New(), Source: uuid.New(), N: 3, Card: uuid.New()},
		members: []entryBatchMember{{
			cardID: uuid.New(), controller: uuid.New(), token: true, attacking: uuid.New(),
			tokenEntry: TokenEntryOptions{Tapped: true, Counters: map[string]int{CounterPlusOne: 1}, Keywords: []string{"haste"}},
		}},
	}
	raw, err := json.Marshal(g.snapshotEntryBatch(b, &ContinuationCensus{}))
	if err != nil {
		t.Fatal(err)
	}
	var back entryBatchSnapshot
	if err := json.Unmarshal(raw, &back); err != nil {
		t.Fatal(err)
	}
	got := restoreEntryBatch(&back)
	if !reflect.DeepEqual(got.members, b.members) || !reflect.DeepEqual(got.attraction, b.attraction) {
		t.Fatalf("got %+v / %+v, want %+v / %+v", got.members, got.attraction, b.members, b.attraction)
	}
}
