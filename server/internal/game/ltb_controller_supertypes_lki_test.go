package game

import (
	"slices"
	"testing"

	"github.com/google/uuid"
)

// ltb_controller_supertypes_lki_test.go — #1682, CR 603.10a: a
// leaving permanent's SUPERTYPES and CONTROLLER as it last existed ride
// its EventLTB (Event.LastKnownSupertypes, Event.LastKnownController),
// beside the types (#1675) and subtypes (#1679). A permanent legendary
// only through an effect is not legendary in the graveyard, and a card
// in a graveyard has no controller (CR 108.4) — so "whenever a
// legendary creature you control dies" reads both off the event.
//
// combatExitRoutes covers every EventLTB emit site.

// legendByEffectOracle is a test oracle whose only static makes the
// permanent itself legendary — a layer-4 supertype grant, the shape a
// Clone copying a legend has on the battlefield. Off the battlefield no
// static applies, so the card is a plain creature again.
const legendByEffectOracle = "test-1682-legendary-by-effect"

func withLegendByEffect(t *testing.T) {
	t.Helper()
	withStaticAbilities(t, func(oracleID string) []StaticAbility {
		if oracleID != legendByEffectOracle {
			return nil
		}
		return []StaticAbility{{
			Layer: Layer4Type,
			AppliesTo: func(target *Card, _ *Game, source *Card) bool {
				return target.InstanceID == source.InstanceID
			},
			Apply: func(c *Characteristic, _ *Card, _ *Game, _ *Card) {
				if !typeListHas(c.Supertypes, "Legendary") {
					c.Supertypes = append(c.Supertypes, "Legendary")
				}
			},
		}}
	})
}

// pushLegendByEffect seeds a non-legendary 2/2 whose own static makes
// it legendary, and checks the layer took.
func pushLegendByEffect(t *testing.T, g *Game, owner *Player) uuid.UUID {
	t.Helper()
	c := NewCard("Test Clone", owner.ID)
	c.TypeLine = "Creature — Shapeshifter"
	c.OracleID = legendByEffectOracle
	c.Power, c.Toughness = 2, 2
	id := pushTypedTestCard(g, c)
	if !layeredBattlefieldCard(t, g, id).HasSupertype("legendary") {
		t.Fatal("setup: the static did not make the creature legendary")
	}
	return id
}

// stealForGameTest hands id to `to` until end of turn and settles the
// layer pass, which is where the control change lands on the card.
func stealForGameTest(t *testing.T, g *Game, id, to uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() {
		if !g.GainControlForEffect(uuid.New(), id, to, g.UntilEndOfTurnDuration(), "test — steal") {
			t.Fatal("setup: GainControlForEffect registered nothing")
		}
		g.RecomputeLayersIfStaleLocked()
	})
	if c, _ := g.battlefieldCardLocked(id); c.Controller != to {
		t.Fatal("setup: the steal did not change the controller")
	}
}

// A creature legendary only through an effect reports that it was on
// every exit route; the card it leaves behind is not legendary.
func TestLTBCarriesLastKnownSupertypesOnEveryExitRoute(t *testing.T) {
	withLegendByEffect(t)
	for _, route := range combatExitRoutes {
		t.Run(route.name, func(t *testing.T) {
			g := newActiveGame(t)
			clone := pushLegendByEffect(t, g, g.Seats[0])

			seq := lastSeq(g)
			route.exit(t, g, clone)
			ev := ltbSince(t, g, seq, clone)
			if !slices.Contains(ev.LastKnownSupertypes, "Legendary") {
				t.Errorf("LastKnownSupertypes = %v, want Legendary", ev.LastKnownSupertypes)
			}
			if was, known := ev.WasSupertype("legendary"); !was || !known {
				t.Errorf("WasSupertype(legendary) = %v, %v; want true, true", was, known)
			}
			if was, known := ev.WasSupertype("Snow"); was || !known {
				t.Errorf("WasSupertype(Snow) = %v, %v; want false, true", was, known)
			}
			if c, ok := g.LookupCardForEffect(clone); ok && c.HasSupertype("legendary") {
				t.Error("the card off the battlefield is still legendary; the event is only needed because it is not")
			}
		})
	}
}

// A stolen creature leaving by any route reports the THIEF as the
// controller it left under — the player who controlled the permanent,
// not the owner of the card.
func TestLTBCarriesLastKnownControllerOnEveryExitRoute(t *testing.T) {
	for _, route := range combatExitRoutes {
		t.Run(route.name, func(t *testing.T) {
			g := newActiveGame(t)
			owner, thief := g.Seats[0], g.Seats[1]
			bear := pushCombatant(t, g, owner, "Grizzly Bears", 2, 2)
			stealForGameTest(t, g, bear, thief.ID)

			seq := lastSeq(g)
			route.exit(t, g, bear)
			ev := ltbSince(t, g, seq, bear)
			if ev.LastKnownController != thief.ID {
				t.Errorf("LastKnownController = %s, want the thief %s (owner %s)", ev.LastKnownController, thief.ID, owner.ID)
			}
			if who, known := ev.LeftUnderControlOf(); who != thief.ID || !known {
				t.Errorf("LeftUnderControlOf = %s, %v; want the thief, true", who, known)
			}
		})
	}
}

// An unstolen permanent's stamp is its owner — the stamp is always a
// real player for a permanent that left, never an unknown.
func TestLTBLastKnownControllerOfAnOwnPermanent(t *testing.T) {
	g := newActiveGame(t)
	owner := g.Seats[0]
	bear := pushCombatant(t, g, owner, "Grizzly Bears", 2, 2)
	seq := lastSeq(g)
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(bear); err != nil {
			t.Fatalf("DestroyPermanentForEffect: %v", err)
		}
	})
	if who, known := ltbSince(t, g, seq, bear).LeftUnderControlOf(); who != owner.ID || !known {
		t.Errorf("LeftUnderControlOf = %s, %v; want the owner, true", who, known)
	}
}

// WasSupertype and LeftUnderControlOf answer "unknown" for anything
// that is not an EventLTB and for an EventLTB with no stamp; a stamped
// event with no supertypes is a known "no".
func TestWasSupertypeAndLeftUnderControlOfAreUnknownWithoutAStamp(t *testing.T) {
	who := uuid.New()
	if _, known := (Event{Kind: EventLTB}).WasSupertype("Legendary"); known {
		t.Error("an unstamped EventLTB claimed to know its supertypes")
	}
	if _, known := (Event{Kind: EventETB, LastKnownTypes: []string{"Creature"}, LastKnownSupertypes: []string{"Legendary"}}).WasSupertype("Legendary"); known {
		t.Error("a non-LTB event claimed last-known supertypes")
	}
	if was, known := (Event{Kind: EventLTB, LastKnownTypes: []string{"Creature"}}).WasSupertype("Legendary"); was || !known {
		t.Errorf("a stamped creature with no supertypes: WasSupertype = %v, %v; want false, true", was, known)
	}
	if _, known := (Event{Kind: EventLTB}).LeftUnderControlOf(); known {
		t.Error("an unstamped EventLTB claimed to know its controller")
	}
	if _, known := (Event{Kind: EventETB, LastKnownController: who}).LeftUnderControlOf(); known {
		t.Error("a non-LTB event claimed a last-known controller")
	}
	if got, known := (Event{Kind: EventLTB, LastKnownController: who}).LeftUnderControlOf(); got != who || !known {
		t.Errorf("a stamped EventLTB: LeftUnderControlOf = %s, %v; want %s, true", got, known, who)
	}
}

// "A creature died under your control this turn" (the per-turn
// CreaturesDied tally, Barrensteppe Siege's Mardu "if") credits the
// controller the event says the creature left under, not the card in
// the graveyard — which, with no controller of its own (CR 108.4),
// names its owner here.
func TestCreaturesDiedTallyCreditsTheLastKnownController(t *testing.T) {
	g := newActiveGame(t)
	owner, thief := g.Seats[0], g.Seats[1]
	card := NewCard("Grizzly Bears", owner.ID)
	card.TypeLine = "Creature — Bear"
	owner.Graveyard.PushTop(card)
	g.WithWriteLock(func() {
		g.EmitEvent(Event{
			Kind: EventLTB, CardID: card.InstanceID, NewZone: ZoneGraveyard,
			LastKnownTypes: []string{"Creature"}, LastKnownController: thief.ID,
		})
	})
	if got := g.TurnTallyFor(thief.ID).CreaturesDied; got != 1 {
		t.Errorf("thief's CreaturesDied = %d, want 1", got)
	}
	if got := g.TurnTallyFor(owner.ID).CreaturesDied; got != 0 {
		t.Errorf("owner's CreaturesDied = %d, want 0 — the creature died under the thief's control", got)
	}
}

// The supertypes and the controller live on the event and nowhere
// else, so the event log carries them through a persisted snapshot, an
// undo clone, and the undo and redo themselves — and a trigger
// context's copy of the event does not share the supertype slice with
// the log.
func TestLTBLastKnownSupertypesAndControllerSurviveSnapshotCloneAndUndo(t *testing.T) {
	withLegendByEffect(t)
	g := newActiveGame(t)
	owner, thief := g.Seats[0], g.Seats[1]
	clone := pushLegendByEffect(t, g, owner)
	stealForGameTest(t, g, clone, thief.ID)
	before := g.Clone()
	seq := lastSeq(g)
	g.WithWriteLock(func() {
		if err := g.SacrificePermanentForEffect(clone); err != nil {
			t.Fatalf("SacrificePermanentForEffect: %v", err)
		}
	})
	want := ltbSince(t, g, seq, clone)
	if !slices.Contains(want.LastKnownSupertypes, "Legendary") || want.LastKnownController != thief.ID {
		t.Fatalf("stamp = %v / %s, want Legendary / the thief", want.LastKnownSupertypes, want.LastKnownController)
	}

	check := func(label string, gg *Game) {
		t.Helper()
		got := ltbSince(t, gg, seq, clone)
		if !slices.Equal(got.LastKnownSupertypes, want.LastKnownSupertypes) {
			t.Errorf("%s: LastKnownSupertypes = %v, want %v", label, got.LastKnownSupertypes, want.LastKnownSupertypes)
		}
		if got.LastKnownController != thief.ID {
			t.Errorf("%s: LastKnownController = %s, want the thief", label, got.LastKnownController)
		}
	}
	_, restored := roundTrip(t, g)
	check("snapshot round trip", restored)
	after := g.Clone()
	check("clone", after)

	g.RestoreFrom(before)
	for _, ev := range g.Events {
		if ev.Seq > seq && ev.Kind == EventLTB && ev.CardID == clone {
			t.Fatalf("undo kept the EventLTB: %+v", ev)
		}
	}
	g.RestoreFrom(after)
	check("restore of the post-sacrifice clone", g)

	tc := &TriggerContext{Event: ltbSince(t, g, seq, clone)}
	cp := cloneTriggerContext(tc)
	cp.Event.LastKnownSupertypes[0] = "Mutated"
	if tc.Event.LastKnownSupertypes[0] == "Mutated" {
		t.Error("cloneTriggerContext shares LastKnownSupertypes with the original")
	}
}
