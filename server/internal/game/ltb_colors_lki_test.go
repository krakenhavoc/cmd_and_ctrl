package game

import (
	"slices"
	"testing"

	"github.com/google/uuid"
)

// ltb_colors_lki_test.go — #1689, CR 603.10a: a leaving permanent's
// COLOURS as it last existed ride its EventLTB (Event.LastKnownColors),
// closing the last-known-information family the #1675 types, #1679
// subtypes and #1682 supertypes/controller amendments opened. A
// permanent black only through an effect (Darkest Hour, a
// black-making static, an Aura) is not black in the graveyard.
//
// combatExitRoutes covers every EventLTB emit site.

// blackByEffectOracle is a test oracle whose only static paints the
// permanent black — a layer-5 colour change (CR 105.3), the shape
// Darkest Hour gives every creature its controller controls.
const blackByEffectOracle = "test-1689-black-by-effect"

func withBlackByEffect(t *testing.T) {
	t.Helper()
	withStaticAbilities(t, func(oracleID string) []StaticAbility {
		if oracleID != blackByEffectOracle {
			return nil
		}
		return []StaticAbility{{
			Layer: Layer5Color,
			AppliesTo: func(target *Card, _ *Game, source *Card) bool {
				return target.InstanceID == source.InstanceID
			},
			Apply: func(c *Characteristic, _ *Card, _ *Game, _ *Card) {
				c.Colors = []string{"B"}
			},
		}}
	})
}

// pushBlackByEffect seeds a printed-green 2/2 whose own static paints
// it black, and checks the layer took.
func pushBlackByEffect(t *testing.T, g *Game, owner *Player) uuid.UUID {
	t.Helper()
	c := NewCard("Test Verdant Beast", owner.ID)
	c.TypeLine = "Creature — Beast"
	c.OracleID = blackByEffectOracle
	c.Colors = []string{"G"}
	c.Power, c.Toughness = 2, 2
	id := pushTypedTestCard(g, c)
	if !layeredBattlefieldCard(t, g, id).HasColor("B") {
		t.Fatal("setup: the static did not paint the creature black")
	}
	return id
}

// A creature black only through an effect reports that it was, on
// every exit route; the card it leaves behind is not black — it is
// the printed green again, because no effect applies in the
// graveyard.
func TestLTBCarriesLastKnownColorsOnEveryExitRoute(t *testing.T) {
	withBlackByEffect(t)
	for _, route := range combatExitRoutes {
		t.Run(route.name, func(t *testing.T) {
			g := newActiveGame(t)
			beast := pushBlackByEffect(t, g, g.Seats[0])

			seq := lastSeq(g)
			route.exit(t, g, beast)
			ev := ltbSince(t, g, seq, beast)
			if !slices.Contains(ev.LastKnownColors, "B") {
				t.Errorf("LastKnownColors = %v, want B", ev.LastKnownColors)
			}
			if was, known := ev.WasColor("B"); !was || !known {
				t.Errorf("WasColor(B) = %v, %v; want true, true", was, known)
			}
			if was, known := ev.WasColor("G"); was || !known {
				t.Errorf("WasColor(G) = %v, %v; want false, true — the layer-5 change REPLACES the colour", was, known)
			}
			if c, ok := g.LookupCardForEffect(beast); ok && c.HasColor("B") {
				t.Error("the card off the battlefield is still black; the event is only needed because it is not")
			}
		})
	}
}

// WasColor answers "unknown" for anything that is not an EventLTB and
// for an EventLTB with no stamp; a stamped event with no colours is a
// known "no" (colourless), never an unknown.
func TestWasColorIsUnknownWithoutAStamp(t *testing.T) {
	if _, known := (Event{Kind: EventLTB}).WasColor("B"); known {
		t.Error("an unstamped EventLTB claimed to know its colours")
	}
	if _, known := (Event{Kind: EventETB, LastKnownTypes: []string{"Creature"}, LastKnownColors: []string{"B"}}).WasColor("B"); known {
		t.Error("a non-LTB event claimed last-known colours")
	}
	if was, known := (Event{Kind: EventLTB, LastKnownTypes: []string{"Creature"}}).WasColor("B"); was || !known {
		t.Errorf("a stamped colourless creature: WasColor = %v, %v; want false, true", was, known)
	}
	if was, known := (Event{Kind: EventLTB, LastKnownTypes: []string{"Creature"}, LastKnownColors: []string{"B"}}).WasColor("B"); !was || !known {
		t.Errorf("a stamped black creature: WasColor = %v, %v; want true, true", was, known)
	}
}

// The colours live on the event and nowhere else, so the event log
// carries them through a persisted snapshot, an undo clone, and the
// undo and redo themselves — and a trigger context's copy of the
// event does not share the colour slice with the log.
func TestLTBLastKnownColorsSurviveSnapshotCloneAndUndo(t *testing.T) {
	withBlackByEffect(t)
	g := newActiveGame(t)
	owner := g.Seats[0]
	beast := pushBlackByEffect(t, g, owner)
	before := g.Clone()
	seq := lastSeq(g)
	g.WithWriteLock(func() {
		if err := g.SacrificePermanentForEffect(beast); err != nil {
			t.Fatalf("SacrificePermanentForEffect: %v", err)
		}
	})
	want := ltbSince(t, g, seq, beast)
	if !slices.Contains(want.LastKnownColors, "B") {
		t.Fatalf("stamp = %v, want B", want.LastKnownColors)
	}

	check := func(label string, gg *Game) {
		t.Helper()
		got := ltbSince(t, gg, seq, beast)
		if !slices.Equal(got.LastKnownColors, want.LastKnownColors) {
			t.Errorf("%s: LastKnownColors = %v, want %v", label, got.LastKnownColors, want.LastKnownColors)
		}
	}
	_, restored := roundTrip(t, g)
	check("snapshot round trip", restored)
	after := g.Clone()
	check("clone", after)

	g.RestoreFrom(before)
	for _, ev := range g.Events {
		if ev.Seq > seq && ev.Kind == EventLTB && ev.CardID == beast {
			t.Fatalf("undo kept the EventLTB: %+v", ev)
		}
	}
	g.RestoreFrom(after)
	check("restore of the post-sacrifice clone", g)

	tc := &TriggerContext{Event: ltbSince(t, g, seq, beast)}
	cp := cloneTriggerContext(tc)
	cp.Event.LastKnownColors[0] = "U"
	if tc.Event.LastKnownColors[0] == "U" {
		t.Error("cloneTriggerContext shares LastKnownColors with the original")
	}
}
