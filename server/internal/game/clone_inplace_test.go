package game

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
)

// clone_inplace_test.go — ADR 0194 option 1: cloneZone copies each
// zone's array once and deepens every card in its slot, instead of
// moving each 1.4 KB Card through cloneCard by value. The result must
// be exactly what the old copy made, and as independent of the live
// game.

// legacyCloneCard is cloneCard as it stood before ADR 0194, kept here
// verbatim as the reference the in-place copy is held to.
func legacyCloneCard(c Card) Card {
	out := c
	if len(c.Keywords) > 0 {
		out.Keywords = append([]string(nil), c.Keywords...)
	}
	if len(c.GrantedAbilities) > 0 {
		out.GrantedAbilities = append([]string(nil), c.GrantedAbilities...)
	}
	if len(c.ManaAbilities) > 0 {
		out.ManaAbilities = append([]ManaAbilityShape(nil), c.ManaAbilities...)
	}
	if len(c.ActivatedAbilities) > 0 {
		out.ActivatedAbilities = append([]ActivatedAbilityShape(nil), c.ActivatedAbilities...)
	}
	if len(c.Colors) > 0 {
		out.Colors = append([]string(nil), c.Colors...)
	}
	if len(c.AlsoBlocking) > 0 {
		out.AlsoBlocking = append([]uuid.UUID(nil), c.AlsoBlocking...)
	}
	if len(c.Faces) > 0 {
		out.Faces = make([]Face, len(c.Faces))
		copy(out.Faces, c.Faces)
		for i := range out.Faces {
			if len(c.Faces[i].Colors) > 0 {
				out.Faces[i].Colors = append([]string(nil), c.Faces[i].Colors...)
			}
		}
	}
	out.PrintedSelf = copyPrintedValues(c.PrintedSelf)
	out.DurationCopyBase = copyPrintedValues(c.DurationCopyBase)
	out.FaceDownListed = c.FaceDownListed.clone()
	out.ModesChosen = copyModesChosen(c.ModesChosen)
	out.CounterStampedAt = copyStringInt64Map(c.CounterStampedAt)
	if len(c.Counters) > 0 {
		out.Counters = make(map[string]int, len(c.Counters))
		for k, v := range c.Counters {
			out.Counters[k] = v
		}
	} else {
		out.Counters = nil
	}
	if len(c.NextUntapSkips) > 0 {
		out.NextUntapSkips = append([]UntapSkip(nil), c.NextUntapSkips...)
	} else {
		out.NextUntapSkips = nil
	}
	out.Goads = cloneGoads(c.Goads)
	out.Provenance = c.Provenance.Clone()
	out.CraftedWith = cloneObjectRefs(c.CraftedWith)
	out.Meld = c.Meld.clone()
	if len(c.MeldedFrom) > 0 {
		out.MeldedFrom = make([]Card, len(c.MeldedFrom))
		for i := range c.MeldedFrom {
			out.MeldedFrom[i] = legacyCloneCard(c.MeldedFrom[i])
		}
	}
	if len(c.KnownBy) > 0 {
		out.KnownBy = make(map[uuid.UUID]bool, len(c.KnownBy))
		for k, v := range c.KnownBy {
			out.KnownBy[k] = v
		}
	} else {
		out.KnownBy = nil
	}
	return out
}

func legacyCloneZone(z *Zone) *Zone {
	if z == nil {
		return nil
	}
	cloned := &Zone{Kind: z.Kind, Owner: z.Owner}
	if len(z.Cards) > 0 {
		cloned.Cards = make([]Card, len(z.Cards))
		for i, c := range z.Cards {
			cloned.Cards[i] = legacyCloneCard(c)
		}
	}
	return cloned
}

// zoneSlots lists every zone pointer a Clone copies, in a fixed order.
func zoneSlots(g *Game) []**Zone {
	out := []**Zone{&g.Battlefield, &g.Stack, &g.Exile, &g.PhasedOut}
	for _, p := range g.Seats {
		out = append(out, &p.Library, &p.Hand, &p.Graveyard, &p.Command, &p.Emblems, &p.Companion, &p.Sideboard)
	}
	return out
}

// richMidgame is midgameBoard with every field cloneCard deepens set on
// some of the cards, nested melded cards included, and an empty but
// non-nil map or slice on others (the copy normalises those to nil).
func richMidgame(tb testing.TB) *Game {
	tb.Helper()
	g := midgameBoard(tb)
	other := g.Seats[1].ID
	g.WithWriteLock(func() {
		bf := g.Battlefield.Cards
		for i := range bf {
			c := &bf[i]
			switch i % 4 {
			case 0:
				c.GrantedAbilities = []string{"hexproof"}
				c.ManaAbilities = []ManaAbilityShape{{}}
				c.ActivatedAbilities = []ActivatedAbilityShape{{Label: "Tap"}}
				c.Colors = []string{"G", "U"}
				c.AlsoBlocking = []uuid.UUID{other}
				c.Faces = []Face{{Name: "Front", Colors: []string{"G"}}, {Name: "Back"}}
				c.PrintedSelf = &PrintedValues{Name: "Copied", Colors: []string{"U"}, Keywords: []string{"flying"}}
				c.DurationCopyBase = &PrintedValues{Name: "Base"}
				c.FaceDownListed = &FaceDownListing{Types: []string{"Creature"}, Subtypes: []string{"Cyberman"}, Power: 2, Toughness: 2}
				c.ModesChosen = map[string][]int{"modal": {0, 2}}
				c.CounterStampedAt = map[string]int64{"flying": 7}
				c.NextUntapSkips = []UntapSkip{{Player: other, Extra: 1}}
				c.Goads = []Goad{{By: other, ExpiresAtTurnsBegun: 3}}
				c.Provenance = CastProvenance{AltCost: "escape", OptionalCosts: []int{1}}
				c.CraftedWith = []ObjectRef{{ID: other, Epoch: 2}}
				c.Meld = &MeldPrint{ResultOracleID: "meld-result"}
				c.KnownBy = map[uuid.UUID]bool{other: true}
				inner := g.Seats[0].Library.Cards[0]
				inner.Counters = map[string]int{"loyalty": 3}
				inner.KnownBy = map[uuid.UUID]bool{other: true}
				inner.Keywords = []string{"vigilance"}
				c.MeldedFrom = []Card{inner, legacyCloneCard(inner)}
			case 1:
				c.Counters = map[string]int{}
				c.KnownBy = map[uuid.UUID]bool{}
				c.NextUntapSkips = []UntapSkip{}
			}
		}
	})
	return g
}

func snapshotBytes(tb testing.TB, g *Game) []byte {
	tb.Helper()
	s := g.CaptureSnapshot()
	s.TakenAt = time.Time{} // the capture's own clock, not the game's
	b, err := json.Marshal(s)
	if err != nil {
		tb.Fatal(err)
	}
	return b
}

// TestCloneZoneInPlaceMatchesTheOldCopy: on a realistic board, the
// in-place copy is the old copy, card for card (reflect.DeepEqual sees
// nil against empty and every nested level) and byte for byte in the
// snapshot encoding.
func TestCloneZoneInPlaceMatchesTheOldCopy(t *testing.T) {
	g := richMidgame(t)
	cards := 0
	for _, slot := range zoneSlots(g) {
		z := *slot
		got, want := cloneZone(z), legacyCloneZone(z)
		if !reflect.DeepEqual(got, want) {
			for i := range want.Cards {
				if !reflect.DeepEqual(got.Cards[i], want.Cards[i]) {
					t.Fatalf("zone %v card %d (%s): in-place copy differs from the old copy", z.Kind, i, want.Cards[i].Name)
				}
			}
			t.Fatalf("zone %v: in-place copy differs from the old copy", z.Kind)
		}
		if z != nil {
			cards += len(z.Cards)
		}
	}
	if cards < 400 {
		t.Fatalf("the board has %d cards, want a realistic midgame of at least 400", cards)
	}

	cloned := g.Clone()
	legacy := g.Clone()
	src := zoneSlots(g)
	for i, slot := range zoneSlots(legacy) {
		*slot = legacyCloneZone(*src[i])
	}
	if a, b := snapshotBytes(t, cloned), snapshotBytes(t, legacy); !bytes.Equal(a, b) {
		t.Fatalf("snapshot of the in-place clone (%d bytes) differs from the old clone's (%d bytes)", len(a), len(b))
	}
}

// mutateEveryDeepenedField writes through every slice, map and pointer
// deepenCard copies, on every card in g's zones.
func mutateEveryDeepenedField(g *Game) {
	who := uuid.New()
	var poke func(c *Card)
	poke = func(c *Card) {
		c.Name += "!"
		if len(c.Keywords) > 0 {
			c.Keywords[0] = "mutated"
		}
		if len(c.GrantedAbilities) > 0 {
			c.GrantedAbilities[0] = "mutated"
		}
		if len(c.ManaAbilities) > 0 {
			c.ManaAbilities[0] = ManaAbilityShape{TapCost: true}
		}
		if len(c.ActivatedAbilities) > 0 {
			c.ActivatedAbilities[0].Label = "mutated"
		}
		if len(c.Colors) > 0 {
			c.Colors[0] = "B"
		}
		if len(c.AlsoBlocking) > 0 {
			c.AlsoBlocking[0] = who
		}
		if len(c.Faces) > 0 {
			c.Faces[0].Name = "mutated"
			if len(c.Faces[0].Colors) > 0 {
				c.Faces[0].Colors[0] = "B"
			}
		}
		if c.PrintedSelf != nil {
			c.PrintedSelf.Name = "mutated"
			c.PrintedSelf.Colors[0] = "B"
		}
		if c.DurationCopyBase != nil {
			c.DurationCopyBase.Name = "mutated"
		}
		if c.FaceDownListed != nil {
			c.FaceDownListed.Types[0] = "Land"
		}
		if c.ModesChosen != nil {
			c.ModesChosen["modal"][0] = 9
			c.ModesChosen["new"] = []int{1}
		}
		if c.CounterStampedAt != nil {
			c.CounterStampedAt["new"] = 1
		}
		if c.Counters != nil {
			c.Counters["mutated"] = 99
		}
		if len(c.NextUntapSkips) > 0 {
			c.NextUntapSkips[0].Extra = 99
		}
		if len(c.Goads) > 0 {
			c.Goads[0].By = who
		}
		if len(c.Provenance.OptionalCosts) > 0 {
			c.Provenance.OptionalCosts[0] = 99
		}
		if len(c.CraftedWith) > 0 {
			c.CraftedWith[0].Epoch = 99
		}
		if c.Meld != nil {
			c.Meld.ResultOracleID = "mutated"
		}
		if c.KnownBy != nil {
			c.KnownBy[who] = true
		}
		for i := range c.MeldedFrom {
			poke(&c.MeldedFrom[i])
		}
	}
	for _, slot := range zoneSlots(g) {
		if z := *slot; z != nil {
			for i := range z.Cards {
				poke(&z.Cards[i])
			}
		}
	}
}

// zoneDiff names the first card field where got and want differ.
func zoneDiff(got, want *Zone) string {
	if got == nil || want == nil || len(got.Cards) != len(want.Cards) {
		return "zone shape"
	}
	for i := range want.Cards {
		g, w := reflect.ValueOf(got.Cards[i]), reflect.ValueOf(want.Cards[i])
		for f := 0; f < g.NumField(); f++ {
			if !g.Field(f).CanInterface() {
				continue
			}
			if !reflect.DeepEqual(g.Field(f).Interface(), w.Field(f).Interface()) {
				return fmt.Sprintf("card %d (%s) field %s", i, want.Cards[i].Name, g.Type().Field(f).Name)
			}
		}
	}
	return "zone"
}

// TestCloneIsIndependentOfTheOriginal: mutating a clone in place never
// changes the game it was cloned from, and mutating the live game
// never changes a clone already taken (the undo ring's guarantee).
func TestCloneIsIndependentOfTheOriginal(t *testing.T) {
	t.Run("mutating the clone", func(t *testing.T) {
		g := richMidgame(t)
		before := snapshotBytes(t, g)
		var want []*Zone
		for _, slot := range zoneSlots(g) {
			want = append(want, legacyCloneZone(*slot))
		}
		clone := g.Clone()
		mutateEveryDeepenedField(clone)
		if !bytes.Equal(before, snapshotBytes(t, g)) {
			t.Fatal("mutating the clone changed the original's snapshot")
		}
		for i, slot := range zoneSlots(g) {
			// Both sides through the same copy, which reads nil and empty alike.
			if got := legacyCloneZone(*slot); !reflect.DeepEqual(got, want[i]) {
				t.Fatalf("mutating the clone changed the original's zone %d: %s", i, zoneDiff(got, want[i]))
			}
		}
	})
	t.Run("mutating the original", func(t *testing.T) {
		g := richMidgame(t)
		clone := g.Clone()
		before := snapshotBytes(t, clone)
		var want []*Zone
		for _, slot := range zoneSlots(clone) {
			want = append(want, legacyCloneZone(*slot))
		}
		g.WithWriteLock(func() { mutateEveryDeepenedField(g) })
		if !bytes.Equal(before, snapshotBytes(t, clone)) {
			t.Fatal("mutating the original changed the clone's snapshot")
		}
		for i, slot := range zoneSlots(clone) {
			if got := legacyCloneZone(*slot); !reflect.DeepEqual(got, want[i]) {
				t.Fatalf("mutating the original changed the clone's zone %d: %s", i, zoneDiff(got, want[i]))
			}
		}
	})
}
