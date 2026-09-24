package game

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"
	"unsafe"

	"github.com/google/uuid"
)

// snapshot_carried_test.go turns `carried` from a promise into a
// property (#1005).
//
// snapshot_drift_test.go makes every domain field say what happens to
// it — carried, rebuilt or dropped — and nothing checked that a field
// SAYING `carried` is carried by anything. The gap is not theoretical
// and it is not visible from either existing test:
//
//   - TestSnapshotRoundTripIsExact compares capture → JSON → restore →
//     capture. Delete a field from BOTH projections and it still
//     passes, because absent equals absent on both sides. It catches a
//     field dropped in ONE direction and nothing else.
//   - the drift test walks field NAMES. It never reads a value.
//   - and the round trip's fixture is `enrich`, so a field the fixture
//     does not set is invisible even in the one-directional case. Three
//     `carried` fields were provably uncovered when #1005 was filed —
//     NamedTribe, ChosenColor and ChosenPlayer, the CR 614.12-family
//     stored answers, none of which enrich was setting. #980 added
//     those three by hand.
//
// So: for every field the plan classifies `carried`, write a
// distinctive value onto a fixture, run the real capture → JSON →
// restore, and read the value back OFF THE RESTORED GAME. A field
// dropped from both projections fails here, and a field ADDED to a plan
// tomorrow is enforced the moment its row lands, with no edit to this
// file — which is the only version of this that survives the four open
// PRs each adding snapshot fields of their own.
//
// Two escape hatches, both narrow and both required to carry a reason:
// carriedFixture, for a field whose type the generator cannot invent a
// plausible value for, and carriedNotRoundTrippable, for the handful
// where capture → restore is genuinely not an identity.

// carriedProbe locates one instance of a planned type inside a game,
// in a way that survives a restore. Every probe is POSITIONAL — the
// first battlefield card, the first seat, the only stack item — because
// a probe that looked its instance up by ID could not test the ID
// field.
type carriedProbe struct {
	typeName string
	plan     fieldPlan

	// capture and restore name the two projection halves in
	// snapshot.go, so a failure can say which one dropped the field.
	capture, restore string

	// at returns the addressable struct value inside g.
	at func(t *testing.T, g *Game) reflect.Value
}

func carriedProbes() []carriedProbe {
	return []carriedProbe{
		{
			typeName: "Game", plan: gameFields,
			capture: "captureSnapshotLocked", restore: "restoreGame",
			at: func(_ *testing.T, g *Game) reflect.Value { return reflect.ValueOf(g).Elem() },
		},
		{
			typeName: "Card", plan: cardFields,
			capture: "snapshotCard", restore: "restoreCard",
			at: func(t *testing.T, g *Game) reflect.Value {
				t.Helper()
				if g.Battlefield == nil || len(g.Battlefield.Cards) == 0 {
					t.Fatal("the fixture has no battlefield card to probe")
				}
				return reflect.ValueOf(&g.Battlefield.Cards[0]).Elem()
			},
		},
		{
			typeName: "Player", plan: playerFields,
			capture: "snapshotPlayer", restore: "restorePlayer",
			at: func(t *testing.T, g *Game) reflect.Value {
				t.Helper()
				if len(g.Seats) == 0 {
					t.Fatal("the fixture has no seat to probe")
				}
				return reflect.ValueOf(g.Seats[0]).Elem()
			},
		},
		{
			typeName: "Zone", plan: zoneFields,
			capture: "snapshotZone", restore: "restoreZone",
			at: func(t *testing.T, g *Game) reflect.Value {
				t.Helper()
				if len(g.Seats) == 0 || g.Seats[0].Graveyard == nil {
					t.Fatal("the fixture has no graveyard to probe")
				}
				return reflect.ValueOf(g.Seats[0].Graveyard).Elem()
			},
		},
		{
			typeName: "StackItem", plan: stackItemFields,
			capture: "snapshotStackItem", restore: "restoreStackItem",
			at: func(t *testing.T, g *Game) reflect.Value {
				t.Helper()
				if len(g.StackMeta) != 1 {
					t.Fatalf("the probe wants exactly one stack item, the fixture has %d", len(g.StackMeta))
				}
				for _, item := range g.StackMeta {
					return reflect.ValueOf(item).Elem()
				}
				return reflect.Value{}
			},
		},
		{
			typeName: "DelayedTrigger", plan: delayedTriggerFields,
			capture: "snapshotDelayedTrigger", restore: "restoreDelayedTrigger",
			at: func(t *testing.T, g *Game) reflect.Value {
				t.Helper()
				if len(g.DelayedTriggers) != 1 {
					t.Fatalf("the probe wants exactly one delayed trigger, the fixture has %d",
						len(g.DelayedTriggers))
				}
				return reflect.ValueOf(g.DelayedTriggers[0]).Elem()
			},
		},
		{
			typeName: "PendingChoice", plan: pendingChoiceFields,
			capture: "snapshotPendingChoice", restore: "restorePendingChoice",
			at: func(t *testing.T, g *Game) reflect.Value {
				t.Helper()
				if len(g.PendingChoices) != 1 {
					t.Fatalf("the probe wants exactly one pending choice, the fixture has %d",
						len(g.PendingChoices))
				}
				return reflect.ValueOf(g.PendingChoices[0]).Elem()
			},
		},
		{
			typeName: "ScopedEffect", plan: scopedEffectFields,
			capture: "captureSnapshotLocked (deepCopyScopedEffects)", restore: "restoreGame (deepCopyScopedEffects)",
			at: func(t *testing.T, g *Game) reflect.Value {
				t.Helper()
				if len(g.ScopedEffects) != 1 {
					t.Fatalf("the probe wants exactly one scoped effect, the fixture has %d", len(g.ScopedEffects))
				}
				return reflect.ValueOf(&g.ScopedEffects[0]).Elem()
			},
		},
		// ScopedStatic has no probe on purpose: the whole entry is
		// `dropped` and censused, so its one `carried` row (Duration) is
		// carried by CLONE and not by the snapshot — see
		// carriedNotRoundTrippable.
	}
}

// carriedFixture holds the hand-written value for a field the generator
// below cannot invent a plausible one for — a typed string with a
// closed set of legal values, an identifier something else validates
// against, a container whose shape another part of the restore
// re-derives. Keyed "<Type>.<Field>".
//
// A value may be the value itself, or a `func(*Game) any` when it has to
// be built against the fixture game (an ID the restore re-stamps from
// somewhere else). The generator fails loudly rather than guessing, so
// an entry here is always the answer to a test that told you exactly
// what it needed.
var carriedFixture = map[string]any{
	// A game's lifecycle state is a closed set (game.go); an invented
	// string would restore a game in a state no code handles.
	"Game.State": StateActive,
	// The zone's own kind, and the one fixture here that is deliberately
	// WRONG for the slot it sits in. restoreZone falls back to the
	// caller's expectation for an empty kind, and every caller's
	// expectation is the slot's real kind — so a graveyard probed with
	// `ZoneGraveyard` would come back right even if both projections
	// dropped the field, and the row would pass forever without
	// checking anything. A graveyard carrying `exile` can only come back
	// as `exile` if something actually carried it.
	"Zone.Kind": ZoneExile,
	// An item's kind decides how the resolution frame routes it.
	"StackItem.Kind": StackItemActivated,
	// The step a delayed trigger waits for.
	"DelayedTrigger.At": StepEnd,
	// The prompt kind, which decides which Resolve* call may answer it.
	"PendingChoice.Kind": PendingChoiceDiscardFromHand,
	// A mod's kind is a closed vocabulary, and restore REFUSES a kind
	// it does not know (ErrUnknownEffectKey, ADR 0041 P4) — so an
	// invented one would fail the restore, not test the carry.
	"ScopedEffect.Mods": []Mod{AddSubtypesMod("drift-ScopedEffect.Mods"), ModifyPTMod(4, 2)},
	"Game.ScopedEffects": func(g *Game) any {
		c := g.Battlefield.Cards[0]
		return []ScopedEffect{{
			Affected:   []AffectedObject{{ID: c.InstanceID, EnteredAt: c.EnteredBattlefieldAt}},
			Mods:       []Mod{SetColorsMod("U")},
			Source:     ObjectRef{ID: uuid.NewSHA1(uuid.Nil, []byte("Game.ScopedEffects")), Epoch: 3},
			SourceName: "drift-Game.ScopedEffects",
			Timestamp:  4343,
			Duration:   Duration{Kind: Indefinite},
			Label:      "drift-Game.ScopedEffects",
		}}
	},
	// The zone a spell was cast from (CR 400.7g / ADR 0066).
	"StackItem.CastFromZone": ZoneGraveyard,
	// ADR 0069's face-down rule. An invented kind has no viewers row.
	"Card.FaceDownKind": FaceDownForetold,

	// The three structural containers. Their ELEMENTS are enforced by
	// the Card / Player / StackItem probes above, so all these rows have
	// to prove is that the container itself comes back — and a generated
	// one cannot, because the restore re-derives part of its shape from
	// somewhere else.

	// restoreGame keys StackMeta by each item's own ID, so a generated
	// map (whose key and whose item.ID are two independent values) can
	// never come back under the key it went in with.
	"Game.StackMeta": func(*Game) any {
		id := uuid.NewSHA1(uuid.Nil, []byte("Game.StackMeta"))
		return map[uuid.UUID]*StackItem{id: {
			ID: id, Kind: StackItemActivated, Label: "drift-Game.StackMeta", Seq: 4242,
		}}
	},
	// restorePlayer re-stamps every zone's owner from the player it is
	// restoring, so a seat has to be built around one ID rather than the
	// generator's three.
	"Game.Seats": func(*Game) any {
		id := uuid.NewSHA1(uuid.Nil, []byte("Game.Seats"))
		return []*Player{{
			ID: id, Name: "drift-Game.Seats", Seat: 0, Life: 4242,
			Library:   newZone(ZoneLibrary, id),
			Hand:      newZone(ZoneHand, id),
			Graveyard: newZone(ZoneGraveyard, id),
			Command:   newZone(ZoneCommand, id),
			Emblems:   newZone(ZoneCommand, id),
			// restorePlayer backfills these three the way clonePlayer and
			// newPlayer leave them, so a fixture that skipped them would
			// fail on the backfill rather than on anything carried.
			CommanderDamage:  map[uuid.UUID]int{},
			CommanderCasts:   map[uuid.UUID]int{},
			LandDropsPerTurn: DefaultLandDropsPerTurn,
		}}
	},
	// #623: restorePlayer stamps the emblem zone's owner from the seat,
	// because a file written before the zone existed restores one owned
	// by nobody. So the fixture owns it the same way a fresh seat does.
	"Player.Emblems": func(g *Game) any {
		z := newZone(ZoneCommand, g.Seats[0].ID)
		z.PushTop(Card{
			InstanceID: uuid.NewSHA1(uuid.Nil, []byte("Player.Emblems")),
			Name:       "drift-Player.Emblems",
			TypeLine:   "Emblem",
			Owner:      g.Seats[0].ID,
			Controller: g.Seats[0].ID,
		})
		return z
	},
}

// carriedNotRoundTrippable names the `carried` fields for which
// capture → restore is deliberately NOT an identity, with the reason and
// the test that covers them instead. Every entry is a claim somebody can
// check, and TestCarriedExemptionsAreLiveAndExplained fails on one that
// no longer names a real field.
var carriedNotRoundTrippable = map[string]string{
	"Game.layerVersion": "restore ADVANCES it by one on purpose, so every restored Card recomputes " +
		"its characteristics instead of serving the cache it dropped (restoreGame). " +
		"TestSnapshotRoundTripIsExact asserts the bump, and TestSnapshotRoundTripKeepsGameUsable " +
		"asserts the recompute it buys.",
	"ScopedStatic.Duration": "ScopedStatic is `dropped` as a whole — it is two closures, counted in " +
		"ContinuationCensus.ScopedStatics — so nothing about it reaches a snapshot. `Duration` is " +
		"marked carried because it is plain data that CLONE carries (undo) and that the snapshot " +
		"could carry the day #515 makes the ability re-derivable. TestUndoKeepsScopedStatics is " +
		"the coverage it has today.",
}

// TestEveryCarriedFieldSurvivesTheSnapshot is the enforcement.
func TestEveryCarriedFieldSurvivesTheSnapshot(t *testing.T) {
	for _, probe := range carriedProbes() {
		for _, name := range carriedFieldsOf(probe.plan) {
			key := probe.typeName + "." + name
			if _, exempt := carriedNotRoundTrippable[key]; exempt {
				continue
			}
			t.Run(key, func(t *testing.T) {
				g := newRestorableGame(t)
				enrich(t, g)

				host := probe.at(t, g)
				field := host.FieldByName(name)
				if !field.IsValid() {
					t.Fatalf("%s is in the plan but not on the struct — "+
						"TestSnapshotCoversEveryDomainField should have said so first", key)
				}
				want := carriedValueFor(t, g, key, field.Type())
				forceSet(field, want)

				snap := g.CaptureSnapshot()
				raw, err := json.Marshal(snap)
				if err != nil {
					t.Fatalf("marshal the snapshot after setting %s: %v", key, err)
				}
				var decoded GameSnapshot
				if err := json.Unmarshal(raw, &decoded); err != nil {
					t.Fatalf("unmarshal the snapshot after setting %s: %v", key, err)
				}
				restored, err := decoded.Restore()
				if err != nil {
					t.Fatalf("restore after setting %s: %v", key, err)
				}

				got := readable(probe.at(t, restored).FieldByName(name))
				if sameValue(t, got, want) {
					return
				}
				t.Errorf(`%s is classified `+"`carried`"+`, but it did not survive
capture → JSON → restore.

  set on the fixture: %s
  read back:          %s

%s%s

Fix the mirror in snapshot.go — it has to be in BOTH halves, because a
field present in one and missing from the other is the only thing the
round-trip property can see — or change the row in
snapshot_drift_test.go to say what really happens to the field.`,
					key, render(want), render(got), firstDifference(want, got),
					blamedProjection(raw, want, probe))
			})
		}
	}
}

// TestCarriedExemptionsAreLiveAndExplained keeps the two escape hatches
// from becoming a place to lose things. An exemption for a field that no
// longer exists is a line the next reader trusts; a fixture for a field
// that no longer needs one is a value nothing checks.
func TestCarriedExemptionsAreLiveAndExplained(t *testing.T) {
	planned := map[string]bool{}
	for _, tc := range driftPlans {
		for _, name := range carriedFieldsOf(tc.plan) {
			planned[driftPlanName(tc.sample)+"."+name] = true
		}
	}

	for key, reason := range carriedNotRoundTrippable {
		if strings.TrimSpace(reason) == "" {
			t.Errorf("carriedNotRoundTrippable[%q] has no reason — an unexplained exemption is a "+
				"field nothing carries and nobody knows it", key)
		}
		if !planned[key] {
			t.Errorf("carriedNotRoundTrippable[%q] exempts a field that is not classified `carried` "+
				"anywhere — it was renamed, removed or reclassified. Delete the entry.", key)
		}
	}
	for key := range carriedFixture {
		if !planned[key] {
			t.Errorf("carriedFixture[%q] names a field that is not classified `carried` anywhere — "+
				"delete it, or the next author will read it as a rule about a live field.", key)
		}
	}
}

// carriedProbeless names the planned types this file deliberately does
// NOT probe, with the reason. Everything else in driftPlans must have a
// probe, so a domain type added to the drift plan cannot arrive with its
// `carried` rows unenforced.
var carriedProbeless = map[string]string{
	"ScopedStatic": "the whole entry is `dropped` and counted in ContinuationCensus.ScopedStatics — " +
		"it never reaches a snapshot at all. Its one `carried` row, Duration, is carried by CLONE; " +
		"see carriedNotRoundTrippable.",
}

// TestEveryPlannedTypeHasACarriedProbe is the half of the enforcement
// that survives somebody adding a domain type rather than a field.
func TestEveryPlannedTypeHasACarriedProbe(t *testing.T) {
	probed := map[string]bool{}
	for _, p := range carriedProbes() {
		probed[p.typeName] = true
	}
	for _, tc := range driftPlans {
		name := driftPlanName(tc.sample)
		if probed[name] {
			continue
		}
		if reason := carriedProbeless[name]; reason != "" {
			continue
		}
		t.Errorf(`%s is classified in snapshot_drift_test.go but has no probe here, so every
row it calls `+"`carried`"+` is a promise again.

Add a carriedProbe for it to carriedProbes() — it needs a positional
way to find one instance of the type in a fixture game, and the names
of its two projection halves in snapshot.go. If the type genuinely
cannot reach a snapshot, add it to carriedProbeless with the reason.`, name)
	}
	for name, reason := range carriedProbeless {
		if strings.TrimSpace(reason) == "" {
			t.Errorf("carriedProbeless[%q] has no reason", name)
		}
		if probed[name] {
			t.Errorf("carriedProbeless[%q] names a type that IS probed — delete the entry", name)
		}
	}
}

// --- the plan walk ---------------------------------------------------

// carriedFieldsOf lists the `carried` rows of a plan, sorted so the
// subtest names are stable.
func carriedFieldsOf(p fieldPlan) []string {
	var out []string
	for name, d := range p {
		if d.how == carried {
			out = append(out, name)
		}
	}
	sort.Strings(out)
	return out
}

// --- building a value ------------------------------------------------

// carriedValueFor is the hand-written fixture for `key` when there is
// one, and a generated distinctive value otherwise.
func carriedValueFor(t *testing.T, g *Game, key string, rt reflect.Type) reflect.Value {
	t.Helper()
	if v, ok := carriedFixture[key]; ok {
		if build, ok := v.(func(*Game) any); ok {
			v = build(g)
		}
		rv := reflect.ValueOf(v)
		if !rv.Type().AssignableTo(rt) {
			if !rv.Type().ConvertibleTo(rt) {
				t.Fatalf("carriedFixture[%q] is a %s, which is neither assignable nor convertible "+
					"to the field's %s", key, rv.Type(), rt)
			}
			rv = rv.Convert(rt)
		}
		if rv.IsZero() {
			t.Fatalf("carriedFixture[%q] is the ZERO value, so the assertion would pass for a field "+
				"nothing carries. Give it a value a restore has to work for.", key)
		}
		return rv
	}
	v, err := generateNonZero(rt, key, 0)
	if err != nil {
		t.Fatalf(`%s needs a hand-written fixture: %v

Add one to carriedFixture in snapshot_carried_test.go, keyed %q, with a
comment saying why the generated value would not do. A field whose type
the generator cannot invent a plausible value for is usually a typed
string with a closed set of legal values, or an identifier something
else validates against.`, key, err, key)
	}
	return v
}

var (
	uuidType = reflect.TypeOf(uuid.UUID{})
	timeType = reflect.TypeOf(time.Time{})
)

// planFor is the drift plan for a domain type, or nil for a type that
// has none (time.Time, Characteristic, the mana token — plain data all
// the way down).
func planFor(rt reflect.Type) fieldPlan {
	switch rt {
	case reflect.TypeOf(Game{}):
		return gameFields
	case reflect.TypeOf(Card{}):
		return cardFields
	case reflect.TypeOf(Player{}):
		return playerFields
	case reflect.TypeOf(Zone{}):
		return zoneFields
	case reflect.TypeOf(StackItem{}):
		return stackItemFields
	case reflect.TypeOf(DelayedTrigger{}):
		return delayedTriggerFields
	case reflect.TypeOf(PendingChoice{}):
		return pendingChoiceFields
	case reflect.TypeOf(ScopedStatic{}):
		return scopedStaticFields
	}
	return nil
}

// generateNonZero builds a distinctive non-zero value of rt. The seed
// makes strings and UUIDs identifiable in the captured JSON, which is
// what lets a failure name the projection that dropped the field.
func generateNonZero(rt reflect.Type, seed string, depth int) (reflect.Value, error) {
	if depth > 4 {
		return reflect.Value{}, fmt.Errorf("type %s nests deeper than the generator goes", rt)
	}
	switch rt {
	case uuidType:
		return reflect.ValueOf(uuid.NewSHA1(uuid.Nil, []byte(seed))), nil
	case timeType:
		return reflect.ValueOf(time.Date(2026, time.September, 19, 12, 0, 0, 0, time.UTC)), nil
	}
	out := reflect.New(rt).Elem()
	switch rt.Kind() {
	case reflect.Bool:
		out.SetBool(true)
	case reflect.String:
		out.SetString("drift-" + seed)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		out.SetInt(4242)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		out.SetUint(4242 % (1 << (rt.Bits() - 1)))
	case reflect.Float32, reflect.Float64:
		out.SetFloat(42.25)
	case reflect.Array:
		for i := 0; i < rt.Len(); i++ {
			elem, err := generateNonZero(rt.Elem(), fmt.Sprintf("%s[%d]", seed, i), depth+1)
			if err != nil {
				return reflect.Value{}, err
			}
			out.Index(i).Set(elem)
		}
	case reflect.Slice:
		elem, err := generateNonZero(rt.Elem(), seed+"[0]", depth+1)
		if err != nil {
			return reflect.Value{}, err
		}
		out.Set(reflect.Append(reflect.MakeSlice(rt, 0, 1), elem))
	case reflect.Map:
		key, err := generateNonZero(rt.Key(), seed+"{key}", depth+1)
		if err != nil {
			return reflect.Value{}, err
		}
		val, err := generateNonZero(rt.Elem(), seed+"{val}", depth+1)
		if err != nil {
			return reflect.Value{}, err
		}
		m := reflect.MakeMap(rt)
		m.SetMapIndex(key, val)
		out.Set(m)
	case reflect.Pointer:
		elem, err := generateNonZero(rt.Elem(), seed+"*", depth+1)
		if err != nil {
			return reflect.Value{}, err
		}
		p := reflect.New(rt.Elem())
		p.Elem().Set(elem)
		out.Set(p)
	case reflect.Struct:
		// A domain type with a plan of its own is filled only in its
		// `carried` fields. Filling a `rebuilt` one (Card.ManaAbilities
		// is two closures re-looked-up from the catalog) would make this
		// test assert the opposite of what that row says.
		inner := planFor(rt)
		filled := false
		for i := 0; i < rt.NumField(); i++ {
			f := rt.Field(i)
			if inner != nil {
				if d, ok := inner[f.Name]; !ok || d.how != carried {
					continue
				}
			}
			elem, err := generateNonZero(f.Type, seed+"."+f.Name, depth+1)
			if err != nil {
				// One field the generator cannot reach does not sink the
				// struct — a distinctive value in any field is enough to
				// prove the field was carried.
				continue
			}
			forceSet(out.Field(i), elem)
			filled = true
		}
		if !filled {
			return reflect.Value{}, fmt.Errorf("no field of %s could be given a value", rt)
		}
	default:
		return reflect.Value{}, fmt.Errorf("the generator has no value for a %s (%s)", rt.Kind(), rt)
	}
	return out, nil
}

// --- reaching unexported fields --------------------------------------
//
// Half of Game's `carried` rows are unexported (the RNG key, the event
// sequence, the combat lock-ins), and reflection refuses to write — or
// even to read as an interface — a field it did not get through an
// exported path. These two go around that, which a TEST may do and
// production code may not: the alternative is leaving exactly the
// fields nobody else can see unenforced.

// forceSet writes v into field, unexported or not.
func forceSet(field, v reflect.Value) {
	if field.CanSet() {
		field.Set(v)
		return
	}
	reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).Elem().Set(v)
}

// readable returns field as a value that can be passed to
// reflect.DeepEqual and json.Marshal, unexported or not.
func readable(field reflect.Value) reflect.Value {
	if field.CanInterface() {
		return field
	}
	return reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).Elem()
}

// --- comparing and blaming -------------------------------------------

// sameValue compares what was written with what came back. JSON rather
// than reflect.DeepEqual, for two reasons that are the same reason: it
// ignores pointer identity, which a restore never preserves, and it
// ignores unexported fields INSIDE the domain types, every one of which
// is a derived cache the plan classifies `rebuilt` (Card.effective,
// StackItem.targetSpec) rather than state a restore owes anybody.
func sameValue(t *testing.T, got, want reflect.Value) bool {
	t.Helper()
	if reflect.DeepEqual(got.Interface(), want.Interface()) {
		return true
	}
	a, errA := json.Marshal(got.Interface())
	b, errB := json.Marshal(want.Interface())
	if errA != nil || errB != nil {
		return false
	}
	// A type whose JSON is contentless (sync/atomic's counters marshal to
	// `{}`) would make this fallback pass for any two values, so those
	// are DeepEqual's alone.
	if len(b) <= 2 || string(b) == "null" {
		return false
	}
	return bytes.Equal(a, b)
}

// blamedProjection says which half of the mirror lost the value, by
// looking for the distinctive value in the CAPTURED json. A value that
// never reached the file was dropped on the way in; one that did was
// dropped on the way out.
func blamedProjection(captured []byte, want reflect.Value, probe carriedProbe) string {
	marker, err := json.Marshal(want.Interface())
	if err == nil && len(marker) > 4 && bytes.Contains(captured, marker) {
		return fmt.Sprintf("The value IS in the captured snapshot, so the RESTORE projection "+
			"dropped it:\n  %s (snapshot.go) does not read it back onto the %s.",
			probe.restore, probe.typeName)
	}
	if err == nil && len(marker) > 4 {
		return fmt.Sprintf("The value is NOT in the captured snapshot, so the CAPTURE projection "+
			"dropped it:\n  %s (snapshot.go) does not put it into the %s mirror.",
			probe.capture, probe.typeName)
	}
	return fmt.Sprintf("The value is not distinctive enough to say which half lost it. Check "+
		"BOTH:\n  %s and %s (snapshot.go).", probe.capture, probe.restore)
}

// firstDifference points at where two long renderings part company, so
// a container field's failure says which of its members went missing
// rather than printing two walls of JSON that start the same.
func firstDifference(want, got reflect.Value) string {
	a, errA := json.Marshal(want.Interface())
	b, errB := json.Marshal(got.Interface())
	if errA != nil || errB != nil || bytes.Equal(a, b) {
		return ""
	}
	i := 0
	for i < len(a) && i < len(b) && a[i] == b[i] {
		i++
	}
	window := func(raw []byte) string {
		end := i + 90
		if end > len(raw) {
			end = len(raw)
		}
		return string(raw[i:end])
	}
	return fmt.Sprintf("  they part company at byte %d:\n    set:  …%s\n    back: …%s\n",
		i, window(a), window(b))
}

// render prints a value the way a failure wants to read it.
func render(v reflect.Value) string {
	if !v.IsValid() {
		return "<invalid>"
	}
	if raw, err := json.Marshal(v.Interface()); err == nil {
		s := string(raw)
		if len(s) > 240 {
			return s[:240] + "…"
		}
		return s
	}
	return fmt.Sprintf("%v", v.Interface())
}
