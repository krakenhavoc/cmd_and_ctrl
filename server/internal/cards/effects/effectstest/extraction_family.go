package effectstest

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// extraction_family_test.go — #4033: Surgical Extraction, Extirpate,
// Eradicate, Scour, Splinter, Sowing Salt and Quash on the graveyard, hand
// and library search. "Any number" leaves every match to the caster; "all"
// finds every match in the graveyard and leaves the hidden zones' matches
// to the caster (CR 701.23b).

const (
	SurgicalExtractionOracle = "3bfa088c-9a47-4123-aee5-4de44e355bef"
	ExtirpateOracle          = "fa857e9d-dfe4-4348-9c1f-16aa240d69cd"
	EradicateOracle          = "782ff740-ff43-4c07-af0b-433cb9770661"
	ScourOracle              = "f18622da-81f3-4854-a1bb-8d658751d8db"
	SplinterOracle           = "90f01e9f-68f1-4e41-97dc-8de990a91504"
	SowingSaltOracle         = "131d2b70-959d-4814-b0c8-38031f141718"
	QuashOracle              = "255131a2-616c-4c8d-a36a-14e5b5ab345e"
)

type ExtractionTable struct {
	G             *game.Game
	Me, Opp, Opp2 *game.Player
	Cards         map[string]uuid.UUID
}

// extractionSetUp gives the opponent `name` twice in the graveyard, once in
// the hand and once in the library, beside a card of another name in each
// zone and a basic Forest in the graveyard.
func ExtractionSetUp(t *testing.T, name, typeLine string) ExtractionTable {
	t.Helper()
	g := NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	x := ExtractionTable{
		G: g, Me: g.Seats[seat], Opp: g.Seats[(seat+1)%len(g.Seats)], Opp2: g.Seats[(seat+2)%len(g.Seats)],
		Cards: map[string]uuid.UUID{},
	}
	g.WithWriteLock(func() { x.Opp.Hand.Cards = nil })
	x.Put("gy", x.Opp.Graveyard, name, typeLine)
	x.Put("gy 2", x.Opp.Graveyard, name, typeLine)
	x.Put("hand", x.Opp.Hand, name, typeLine)
	x.Put("lib", x.Opp.Library, name, typeLine)
	x.Put("gy other", x.Opp.Graveyard, "Shock", "Instant")
	x.Put("hand other", x.Opp.Hand, "Counterspell", "Instant")
	x.Put("lib other", x.Opp.Library, "Opt", "Instant")
	x.Put("gy forest", x.Opp.Graveyard, "Forest", "Basic Land — Forest")
	return x
}

func (x ExtractionTable) Put(key string, zone *game.Zone, name, typeLine string) {
	id := uuid.New()
	zone.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, Owner: x.Opp.ID, Controller: x.Opp.ID})
	x.Cards[key] = id
}

// permanent puts a permanent of `name` onto the battlefield under the
// opponent's control.
func (x ExtractionTable) Permanent(key, name, typeLine string, colors ...string) uuid.UUID {
	id := PushBattlefieldCardWithTimestamp(x.G, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, Colors: colors,
		Power: 2, Toughness: 2, Owner: x.Opp.ID, Controller: x.Opp.ID,
	})
	x.Cards[key] = id
	return id
}

func (x ExtractionTable) Exiled(t *testing.T, keys ...string) {
	t.Helper()
	for _, k := range keys {
		if !x.G.Exile.Contains(x.Cards[k]) {
			t.Errorf("%s was not exiled", k)
		}
	}
}

func (x ExtractionTable) NotExiled(t *testing.T, keys ...string) {
	t.Helper()
	for _, k := range keys {
		if x.G.Exile.Contains(x.Cards[k]) {
			t.Errorf("%s was exiled", k)
		}
	}
}

// searches counts the searches of the opponent's library that shuffled it.
func (x ExtractionTable) Searches() int {
	n := 0
	for _, ev := range x.G.Events {
		if ev.Kind == game.EventSearchLibrary && ev.Target == x.Opp.ID && ev.Label == game.SearchShuffledLabel {
			n++
		}
	}
	return n
}

func (x ExtractionTable) Target(key string) []game.TargetRef {
	return []game.TargetRef{{Kind: game.TargetCard, ID: x.Cards[key]}}
}

// candidates is the open search prompt's candidate set, by key.
func (x ExtractionTable) Candidates(t *testing.T) []string {
	t.Helper()
	c := SearchChoiceFor(x.G, x.Me.ID)
	if c == nil {
		t.Fatal("no search prompt for the caster")
	}
	var keys []string
	for k, id := range x.Cards {
		if slices.Contains(c.SearchCards, id) {
			keys = append(keys, k)
		}
	}
	slices.Sort(keys)
	return keys
}

func WantKeys(t *testing.T, got []string, want ...string) {
	t.Helper()
	slices.Sort(want)
	if !slices.Equal(got, want) {
		t.Errorf("candidates %v, want %v", got, want)
	}
}
