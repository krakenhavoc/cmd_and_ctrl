package game

import (
	"slices"
	"testing"

	"github.com/google/uuid"
)

// ltb_keywords_lki_test.go — #4613, CR 603.10a: a leaving permanent's
// KEYWORDS as it last existed ride its EventLTB (Event.LastKnownKeywords),
// the sibling of the #1689 colours. A creature with deathtouch only through
// an effect is a plain creature in the graveyard, and The Serpent Society's
// "another creature you control with deathtouch dies" is judged on what it
// was as it left.

// deathtouchByEffectOracle is a test oracle whose only static gives the
// permanent deathtouch — a layer-6 ability grant (CR 613.1f).
const deathtouchByEffectOracle = "test-4613-deathtouch-by-effect"

func withDeathtouchByEffect(t *testing.T) {
	t.Helper()
	withStaticAbilities(t, func(oracleID string) []StaticAbility {
		if oracleID != deathtouchByEffectOracle {
			return nil
		}
		return []StaticAbility{{
			Layer: Layer6Ability,
			AppliesTo: func(target *Card, _ *Game, source *Card) bool {
				return target.InstanceID == source.InstanceID
			},
			Apply: func(c *Characteristic, _ *Card, _ *Game, _ *Card) {
				c.Abilities = append(c.Abilities, "deathtouch")
			},
		}}
	})
}

// pushDeathtouchByEffect seeds a vanilla 2/2 whose own static gives it
// deathtouch, and checks the layer took.
func pushDeathtouchByEffect(t *testing.T, g *Game, owner *Player) uuid.UUID {
	t.Helper()
	c := NewCard("Test Venomous Beast", owner.ID)
	c.TypeLine = "Creature — Beast"
	c.OracleID = deathtouchByEffectOracle
	c.Power, c.Toughness = 2, 2
	id := pushTypedTestCard(g, c)
	if lc := layeredBattlefieldCard(t, g, id); !HasKeyword(&lc, "deathtouch") {
		t.Fatal("setup: the static did not give the creature deathtouch")
	}
	return id
}

// A creature with deathtouch only through an effect reports that it had it,
// on every exit route; the card it leaves behind has none.
func TestLTBCarriesLastKnownKeywordsOnEveryExitRoute(t *testing.T) {
	withDeathtouchByEffect(t)
	for _, route := range combatExitRoutes {
		t.Run(route.name, func(t *testing.T) {
			g := newActiveGame(t)
			beast := pushDeathtouchByEffect(t, g, g.Seats[0])

			seq := lastSeq(g)
			route.exit(t, g, beast)
			ev := ltbSince(t, g, seq, beast)
			if !slices.Contains(ev.LastKnownKeywords, "deathtouch") {
				t.Errorf("LastKnownKeywords = %v, want deathtouch", ev.LastKnownKeywords)
			}
			if was, known := ev.WasKeyword("deathtouch"); !was || !known {
				t.Errorf("WasKeyword(deathtouch) = %v, %v; want true, true", was, known)
			}
			if was, known := ev.WasKeyword("flying"); was || !known {
				t.Errorf("WasKeyword(flying) = %v, %v; want false, true", was, known)
			}
			if c, ok := g.LookupCardForEffect(beast); ok && HasKeyword(&c, "deathtouch") {
				t.Error("the card off the battlefield still has deathtouch; the event is only needed because it does not")
			}
		})
	}
}

// WasKeyword answers "unknown" for anything that is not an EventLTB and for
// an EventLTB with no stamp; a stamped event with no keywords is a known
// "no", never an unknown.
func TestWasKeywordIsUnknownWithoutAStamp(t *testing.T) {
	if _, known := (Event{Kind: EventLTB}).WasKeyword("deathtouch"); known {
		t.Error("an unstamped EventLTB claimed to know its keywords")
	}
	if _, known := (Event{Kind: EventETB, LastKnownTypes: []string{"Creature"}, LastKnownKeywords: []string{"deathtouch"}}).WasKeyword("deathtouch"); known {
		t.Error("a non-LTB event claimed last-known keywords")
	}
	if was, known := (Event{Kind: EventLTB, LastKnownTypes: []string{"Creature"}}).WasKeyword("deathtouch"); was || !known {
		t.Errorf("a stamped creature with none: WasKeyword = %v, %v; want false, true", was, known)
	}
	if was, known := (Event{Kind: EventLTB, LastKnownTypes: []string{"Creature"}, LastKnownKeywords: []string{"deathtouch"}}).WasKeyword("deathtouch"); !was || !known {
		t.Errorf("a stamped deathtouch creature: WasKeyword = %v, %v; want true, true", was, known)
	}
}

// The keywords live on the event and nowhere else, so the event log carries
// them through a persisted snapshot and an undo clone, and a trigger
// context's copy of the event does not share the slice with the log.
func TestLTBLastKnownKeywordsSurviveSnapshotCloneAndUndo(t *testing.T) {
	withDeathtouchByEffect(t)
	g := newActiveGame(t)
	beast := pushDeathtouchByEffect(t, g, g.Seats[0])
	seq := lastSeq(g)
	g.WithWriteLock(func() {
		if err := g.SacrificePermanentForEffect(beast); err != nil {
			t.Fatalf("SacrificePermanentForEffect: %v", err)
		}
	})
	want := ltbSince(t, g, seq, beast)
	if !slices.Contains(want.LastKnownKeywords, "deathtouch") {
		t.Fatalf("stamp = %v, want deathtouch", want.LastKnownKeywords)
	}

	check := func(label string, gg *Game) {
		t.Helper()
		got := ltbSince(t, gg, seq, beast)
		if !slices.Equal(got.LastKnownKeywords, want.LastKnownKeywords) {
			t.Errorf("%s: LastKnownKeywords = %v, want %v", label, got.LastKnownKeywords, want.LastKnownKeywords)
		}
	}
	_, restored := roundTrip(t, g)
	check("snapshot round trip", restored)
	check("clone", g.Clone())

	tc := &TriggerContext{Event: ltbSince(t, g, seq, beast)}
	cp := cloneTriggerContext(tc)
	cp.Event.LastKnownKeywords[0] = "flying"
	if tc.Event.LastKnownKeywords[0] == "flying" {
		t.Error("cloneTriggerContext shares LastKnownKeywords with the original")
	}
}
