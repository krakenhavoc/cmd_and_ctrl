package game

import (
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/google/uuid"
)

// scoped_effects_test.go pins ADR 0041 phase 3's data record (#1497):
// every mod kind does what its constructor says in the layer it
// names, a record is a restore point, and a restore point naming an
// effect key this binary cannot interpret is refused rather than
// guessed at.

// scopedEffectChar recomputes and returns the card's post-layer
// characteristic.
func scopedEffectChar(t *testing.T, g *Game, id uuid.UUID) Characteristic {
	t.Helper()
	var out Characteristic
	found := false
	g.WithWriteLock(func() {
		g.RecomputeLayersIfStaleLocked()
		if c, ok := g.battlefieldCardLocked(id); ok {
			out, found = c.Effective(), true
		}
	})
	if !found {
		t.Fatalf("card %s not on battlefield", id)
	}
	return out
}

func registerScopedEffectForTest(t *testing.T, g *Game, id uuid.UUID, mods []Mod, d Duration) {
	t.Helper()
	ok := false
	g.WithWriteLock(func() {
		ok = g.RegisterScopedEffectForEffect(uuid.Nil, g.PinnedObjectsLocked(id), mods, d, "test")
	})
	if !ok {
		t.Fatal("RegisterScopedEffectForEffect registered nothing")
	}
}

func TestEveryModKindAppliesInItsLayer(t *testing.T) {
	cases := []struct {
		name  string
		mods  []Mod
		check func(t *testing.T, before, after Characteristic, g *Game)
	}{
		{"setController", nil, nil}, // filled in below: needs a player ID
		{"addTypes", []Mod{AddTypesMod("Artifact")}, func(t *testing.T, _, c Characteristic, _ *Game) {
			if !typeListHas(c.Types, "Artifact") || !typeListHas(c.Types, "Creature") {
				t.Errorf("types = %v, want Artifact added to Creature", c.Types)
			}
		}},
		{"removeTypes", []Mod{RemoveTypesMod("Creature")}, func(t *testing.T, _, c Characteristic, _ *Game) {
			if typeListHas(c.Types, "Creature") {
				t.Errorf("types = %v, want Creature removed", c.Types)
			}
		}},
		{"addSubtypes", []Mod{AddSubtypesMod("Orc")}, func(t *testing.T, _, c Characteristic, _ *Game) {
			if !typeListHas(c.Subtypes, "Orc") || !typeListHas(c.Subtypes, "Bear") {
				t.Errorf("subtypes = %v, want Orc added to Bear", c.Subtypes)
			}
		}},
		{"allCreatureTypes", []Mod{AllCreatureTypesMod()}, func(t *testing.T, _, c Characteristic, _ *Game) {
			if !c.AllCreatureTypes {
				t.Error("AllCreatureTypes is not set")
			}
		}},
		{"setColors", []Mod{SetColorsMod("U")}, func(t *testing.T, _, c Characteristic, _ *Game) {
			if !reflect.DeepEqual(c.Colors, []string{"U"}) {
				t.Errorf("colors = %v, want [U]", c.Colors)
			}
		}},
		{"addKeywords", []Mod{AddKeywordsMod("flying", "flying")}, func(t *testing.T, _, c Characteristic, _ *Game) {
			n := 0
			for _, a := range c.Abilities {
				if a == "flying" {
					n++
				}
			}
			if n != 1 {
				t.Errorf("abilities = %v, want flying exactly once", c.Abilities)
			}
		}},
		{"removeKeywords", []Mod{AddKeywordsMod("hexproof"), RemoveKeywordsMod("hexproof")}, func(t *testing.T, _, c Characteristic, _ *Game) {
			for _, a := range c.Abilities {
				if a == "hexproof" {
					t.Errorf("abilities = %v, want hexproof removed", c.Abilities)
				}
			}
		}},
		{"loseAllAbilities", []Mod{AddKeywordsMod("trample"), LoseAllAbilitiesMod("defender")}, func(t *testing.T, _, c Characteristic, _ *Game) {
			if !reflect.DeepEqual(c.Abilities, []string{"defender"}) {
				t.Errorf("abilities = %v, want only the kept defender", c.Abilities)
			}
			if !c.AbilitiesRemoved {
				t.Error("AbilitiesRemoved is not stamped")
			}
		}},
		{"addRestrictions", []Mod{AddRestrictionsMod(CantBlock)}, func(t *testing.T, _, c Characteristic, _ *Game) {
			if c.Restrictions&CantBlock == 0 {
				t.Errorf("restrictions = %v, want CantBlock", c.Restrictions)
			}
		}},
		{"setBasePT", SetBasePTMods(0, 5), func(t *testing.T, _, c Characteristic, _ *Game) {
			if c.Power != 0 || c.Toughness != 5 {
				t.Errorf("P/T = %d/%d, want 0/5", c.Power, c.Toughness)
			}
		}},
		{"modifyPT", []Mod{ModifyPTMod(3, -1)}, func(t *testing.T, b, c Characteristic, _ *Game) {
			if c.Power != b.Power+3 || c.Toughness != b.Toughness-1 {
				t.Errorf("P/T = %d/%d, want %d/%d", c.Power, c.Toughness, b.Power+3, b.Toughness-1)
			}
		}},
		// #1571: one requirement per mod, attributed to the source.
		{"addAttackRequirement", []Mod{AddAttackRequirementMod(uuid.Nil), AddAttackRequirementMod(attackRequirementTestPlayer)}, func(t *testing.T, _, c Characteristic, _ *Game) {
			if len(c.AttackRequirements) != 2 || c.AttackRequirements[0].OtherThan != uuid.Nil ||
				c.AttackRequirements[1].OtherThan != attackRequirementTestPlayer {
				t.Errorf("attack requirements = %+v, want a plain one and one naming the player", c.AttackRequirements)
			}
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := newActiveGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			id := pushScopedTestCreature(g, me.ID, 2, 2)
			mods, check := tc.mods, tc.check
			if tc.name == "setController" {
				mods = []Mod{SetControllerMod(opp.ID)}
				check = func(t *testing.T, _, c Characteristic, _ *Game) {
					if c.Controller != opp.ID {
						t.Errorf("controller = %s, want %s", c.Controller, opp.ID)
					}
				}
			}
			before := scopedEffectChar(t, g, id)
			registerScopedEffectForTest(t, g, id, mods, IndefiniteDuration())
			check(t, before, scopedEffectChar(t, g, id), g)
		})
	}
}

// attackRequirementTestPlayer is the "other than" player the
// addAttackRequirement case names — any ID will do, the mod only
// carries it.
var attackRequirementTestPlayer = uuid.MustParse("00000000-0000-0000-0000-000000001571")

// TestEveryModKindHasATestCase keeps the table above honest: a kind
// added to the vocabulary without a case is a kind nothing checks.
func TestEveryModKindHasATestCase(t *testing.T) {
	covered := map[ModKind]bool{
		ModSetController: true, ModAddTypes: true, ModRemoveTypes: true, ModAddSubtypes: true,
		ModAllCreatureTypes: true, ModSetColors: true, ModAddKeywords: true, ModRemoveKeywords: true,
		ModLoseAllAbilities: true, ModAddRestrictions: true, ModSetBasePower: true,
		ModSetBaseToughness: true, ModModifyPT: true, ModAddAttackRequirement: true,
	}
	for _, k := range ModKinds() {
		if !covered[k] {
			t.Errorf("mod kind %q has no case in TestEveryModKindAppliesInItsLayer", k)
		}
	}
}

func TestRegisteringAnUnknownModKindPanics(t *testing.T) {
	g := newActiveGame(t)
	id := pushScopedTestCreature(g, g.Seats[0].ID, 2, 2)
	defer func() {
		if recover() == nil {
			t.Error("an unknown mod kind registered without a panic")
		}
	}()
	g.WithWriteLock(func() {
		g.RegisterScopedEffectForEffect(uuid.Nil, g.PinnedObjectsLocked(id),
			[]Mod{{Kind: "notAKind"}}, IndefiniteDuration(), "test")
	})
}

// TestAScopedEffectIsARestorePoint is the point of the slice: a game
// holding only data-backed effects is a full restore point, and the
// restored game applies them.
func TestAScopedEffectIsARestorePoint(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	id := pushScopedTestCreature(g, me.ID, 2, 2)
	registerScopedEffectForTest(t, g, id, []Mod{ModifyPTMod(2, 2), AddSubtypesMod("Orc")}, IndefiniteDuration())

	snap := g.CaptureSnapshot()
	if !snap.Restorable() {
		t.Fatalf("a game holding only a scoped effect is not a restore point: %+v", snap.Continuations)
	}
	raw, err := json.Marshal(snap)
	if err != nil {
		t.Fatal(err)
	}
	var decoded GameSnapshot
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatal(err)
	}
	restored, err := decoded.RestoreStrict()
	if err != nil {
		t.Fatalf("RestoreStrict: %v", err)
	}
	c := scopedEffectChar(t, restored, id)
	if c.Power != 4 || c.Toughness != 4 || !typeListHas(c.Subtypes, "Orc") {
		t.Errorf("restored: %d/%d subtypes %v, want 4/4 with Orc", c.Power, c.Toughness, c.Subtypes)
	}
}

// corruptAndRestore captures a game with one scoped effect, lets the
// caller damage the snapshot, and restores it both ways.
func corruptAndRestore(t *testing.T, corrupt func(s *GameSnapshot)) (error, error) {
	t.Helper()
	g := newActiveGame(t)
	id := pushScopedTestCreature(g, g.Seats[0].ID, 2, 2)
	registerScopedEffectForTest(t, g, id, []Mod{ModifyPTMod(1, 1)}, IndefiniteDuration())
	g.WithWriteLock(func() {
		g.DelayedTriggers = []*DelayedTrigger{{ID: uuid.New(), Controller: g.Seats[0].ID, At: StepEnd,
			Body: carriedTestBodyKey}}
	})
	snap := g.CaptureSnapshot()
	corrupt(snap)
	_, errLoose := snap.Restore()
	_, errStrict := snap.RestoreStrict()
	return errLoose, errStrict
}

// TestAnUnknownEffectKeyIsRefused is ADR 0041 P4's promise: a file
// naming a mod kind, or a tier-2 body or condition key, that this
// binary cannot interpret is the rollback case, and is refused with
// ErrUnknownEffectKey — by the loose restore and the strict one — not
// restored with the effect silently missing.
func TestAnUnknownEffectKeyIsRefused(t *testing.T) {
	cases := map[string]func(s *GameSnapshot){
		"mod kind": func(s *GameSnapshot) { s.ScopedEffects[0].Mods[0].Kind = "grantAbilitiesFromTheFuture" },
		"delayed-trigger body": func(s *GameSnapshot) {
			s.DelayedTriggers[0].Body = "nobody/registered-this-body"
		},
		"delayed-trigger condition": func(s *GameSnapshot) {
			s.DelayedTriggers[0].Condition = "nobody/registered-this-condition"
		},
		"stack-item body": func(s *GameSnapshot) {
			s.PendingTriggers = append(s.PendingTriggers, stackItemSnapshot{ID: uuid.New(), Body: "nobody/registered-this-body"})
		},
	}
	for name, corrupt := range cases {
		t.Run(name, func(t *testing.T) {
			loose, strict := corruptAndRestore(t, corrupt)
			for which, err := range map[string]error{"Restore": loose, "RestoreStrict": strict} {
				if !errors.Is(err, ErrUnknownEffectKey) {
					t.Errorf("%s: err = %v, want ErrUnknownEffectKey", which, err)
				}
			}
		})
	}
	// And the uncorrupted file is fine.
	loose, strict := corruptAndRestore(t, func(*GameSnapshot) {})
	if loose != nil || strict != nil {
		t.Errorf("an intact file was refused: %v / %v", loose, strict)
	}
}

// TestScopedEffectIsPureData holds ADR 0041 P1's parameter rule: the
// record, and everything it holds, is bool, ints, strings, UUIDs and
// the engine's enums — no func, interface, pointer or map.
func TestScopedEffectIsPureData(t *testing.T) {
	var walk func(rt reflect.Type, path string)
	seen := map[reflect.Type]bool{}
	walk = func(rt reflect.Type, path string) {
		if seen[rt] {
			return
		}
		seen[rt] = true
		switch rt.Kind() {
		case reflect.Func, reflect.Interface, reflect.Ptr, reflect.Map, reflect.Chan, reflect.UnsafePointer:
			t.Errorf("%s is a %s; a ScopedEffect holds plain data only", path, rt.Kind())
		case reflect.Slice, reflect.Array:
			if rt == reflect.TypeOf(uuid.UUID{}) {
				return
			}
			walk(rt.Elem(), path+"[]")
		case reflect.Struct:
			for i := 0; i < rt.NumField(); i++ {
				f := rt.Field(i)
				if f.PkgPath != "" {
					t.Errorf("%s.%s is unexported and would not reach the snapshot", path, f.Name)
				}
				walk(f.Type, path+"."+f.Name)
			}
		}
	}
	walk(reflect.TypeOf(ScopedEffect{}), "ScopedEffect")
}

// TestAmassedArmyIsARestorePoint: before #1497 the Orc on a Zombie
// Army was a closure for the rest of the game, and the table had no
// restore point for as long as the Army lived.
func TestAmassedArmyIsARestorePoint(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	existing := seedArmy(g, me.ID, "Zombie")
	amass(t, g, me, "Orc", 1)

	snap := g.CaptureSnapshot()
	if !snap.Restorable() {
		t.Fatalf("an amassed Army is not a restore point: %+v", snap.Continuations)
	}
	restored, err := snap.RestoreStrict()
	if err != nil {
		t.Fatalf("RestoreStrict: %v", err)
	}
	c := scopedEffectChar(t, restored, existing)
	if !typeListHas(c.Subtypes, "Orc") || !typeListHas(c.Subtypes, "Zombie") {
		t.Errorf("restored Army subtypes = %v, want Zombie and Orc", c.Subtypes)
	}
	for _, l := range snap.Continuations.Labels {
		if strings.Contains(l, "amass") {
			t.Errorf("census names amass: %q", l)
		}
	}
}

// TestAPinnedRecordSurvivesItsObjectPhasing: the pin is garbage
// collection for an object that is GONE, and a phased-out permanent is
// not gone (CR 702.26d). The record stops applying while the object is
// out — the layer pass cannot see it — and applies again when it
// phases in.
func TestAPinnedRecordSurvivesItsObjectPhasing(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me := g.Seats[0].ID
	bear := pushPhasingTestCard(g, me, "Grizzly Bears", "Creature — Bear")
	var d Duration
	g.WithWriteLock(func() { d = g.PinnedTo(IndefiniteDuration(), bear) })
	registerScopedEffectForTest(t, g, bear, []Mod{AddSubtypesMod("Orc")}, d)

	g.WithWriteLock(func() { _ = g.PhaseOutForEffect(uuid.Nil, bear) })
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked(); g.ClearExpiredScopedStaticsLocked() })
	if n := len(g.ScopedEffects); n != 1 {
		t.Fatalf("%d scoped effects after the phase-out, want 1 — phasing is not a zone change", n)
	}
	g.WithWriteLock(func() { g.performPhasingLocked(me) })
	if c := scopedEffectChar(t, g, bear); !typeListHas(c.Subtypes, "Orc") {
		t.Errorf("after phasing in: subtypes %v, want Orc", c.Subtypes)
	}
}

// TestAForAsLongAsDurationEndsWhenItsSourcePhasesOut is CR 702.26f,
// and the reason only the PIN looks in g.PhasedOut: a "for as long as
// ~ remains on the battlefield" duration tracks its source, and a
// phased-out source can no longer be seen, so the effect ends — a
// Sower of Temptation that phases out gives the creature back.
func TestAForAsLongAsDurationEndsWhenItsSourcePhasesOut(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me, opp := g.Seats[0].ID, g.Seats[1].ID
	sower := pushPhasingTestCard(g, me, "Sower of Temptation", "Creature — Faerie Wizard")
	victim := pushPhasingTestCard(g, opp, "Grizzly Bears", "Creature — Bear")
	g.WithWriteLock(func() {
		d, ok := g.ForAsLongAsOnBattlefieldDuration(sower)
		if !ok {
			t.Fatal("setup: the Sower is not on the battlefield")
		}
		g.GainControlForEffect(sower, victim, me, d, "Sower of Temptation")
	})
	if got := controllerOfCard(t, g, victim); got != me {
		t.Fatalf("setup: controller %s, want %s", got, me)
	}
	g.WithWriteLock(func() { _ = g.PhaseOutForEffect(uuid.Nil, sower) })
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	if got := controllerOfCard(t, g, victim); got != opp {
		t.Errorf("the Sower phased out and the theft survived: controller %s, want %s (CR 702.26f)", got, opp)
	}
	if n := len(g.ScopedEffects); n != 0 {
		t.Errorf("%d scoped effects survive their source phasing out, want 0", n)
	}
}

// TestAnUnknownDurationIsRefused is the #1497 review's second finding.
// A duration's kind and condition are bare ints on disk, so a kind or
// a condition a newer build adds would otherwise restore as an effect
// that never ends, or fall through to "source on battlefield". Both
// are the rollback case, and both are refused.
func TestAnUnknownDurationIsRefused(t *testing.T) {
	cases := map[string]func(s *GameSnapshot){
		"scoped-effect duration kind": func(s *GameSnapshot) {
			s.ScopedEffects[0].Duration.Kind = DurationKind(99)
		},
		"scoped-effect duration condition": func(s *GameSnapshot) {
			s.ScopedEffects[0].Duration.Condition = DurationCondition(42)
		},
		"delayed-trigger duration kind": func(s *GameSnapshot) {
			d := Duration{Kind: DurationKind(99)}
			s.DelayedTriggers[0].Duration = &d
		},
	}
	for name, corrupt := range cases {
		t.Run(name, func(t *testing.T) {
			loose, strict := corruptAndRestore(t, corrupt)
			for which, err := range map[string]error{"Restore": loose, "RestoreStrict": strict} {
				if !errors.Is(err, ErrUnknownEffectKey) {
					t.Errorf("%s: err = %v, want ErrUnknownEffectKey", which, err)
				}
			}
		})
	}
}

// TestEveryDurationKindAndConditionIsKnown keeps the sentinels honest:
// a kind declared after the sentinel would be refused by the binary
// that declares it.
func TestEveryDurationKindAndConditionIsKnown(t *testing.T) {
	for _, k := range []DurationKind{UntilEndOfTurn, UntilYourNextTurn, ForAsLongAs, Indefinite, WhileInZone} {
		if !k.Known() {
			t.Errorf("duration kind %v is not Known — is it after durationKindEnd?", k)
		}
	}
	for _, c := range []DurationCondition{WhileSourceOnBattlefield, WhileYouControlSource,
		WhileYouControlSourceOnceItLands, WhileSourceRemainsTapped} {
		if !c.Known() {
			t.Errorf("duration condition %d is not Known — is it after durationConditionEnd?", c)
		}
	}
	if durationKindEnd.Known() || durationConditionEnd.Known() {
		t.Error("a sentinel reports itself Known")
	}
}

// TestAnUnknownFieldOnARecordIsRefused: encoding/json drops a key it
// has no field for, and on these four types an unknown key is what a
// newer build's vocabulary looks like — a field a future mod kind
// reads, a duration field a future condition needs. Refused, not
// dropped (ADR 0041 P4, #1497 review).
func TestAnUnknownFieldOnARecordIsRefused(t *testing.T) {
	g := newActiveGame(t)
	id := pushScopedTestCreature(g, g.Seats[0].ID, 2, 2)
	registerScopedEffectForTest(t, g, id, []Mod{ModifyPTMod(1, 1)}, IndefiniteDuration())
	raw, err := json.Marshal(g.CaptureSnapshot())
	if err != nil {
		t.Fatal(err)
	}
	cases := map[string]func(rec map[string]any){
		"record":   func(rec map[string]any) { rec["fromTheFuture"] = true },
		"affected": func(rec map[string]any) { rec["affected"].([]any)[0].(map[string]any)["phaseTag"] = 1 },
		"mod":      func(rec map[string]any) { rec["mods"].([]any)[0].(map[string]any)["counterKind"] = "blight" },
		"duration": func(rec map[string]any) { rec["duration"].(map[string]any)["CounterKind"] = "blight" },
	}
	for name, corrupt := range cases {
		t.Run(name, func(t *testing.T) {
			var generic map[string]any
			if err := json.Unmarshal(raw, &generic); err != nil {
				t.Fatal(err)
			}
			corrupt(generic["scopedEffects"].([]any)[0].(map[string]any))
			bad, err := json.Marshal(generic)
			if err != nil {
				t.Fatal(err)
			}
			var snap GameSnapshot
			if err := json.Unmarshal(bad, &snap); err != nil {
				t.Fatalf("decode: %v", err)
			}
			if _, err := snap.Restore(); !errors.Is(err, ErrUnknownEffectKey) {
				t.Errorf("Restore: err = %v, want ErrUnknownEffectKey", err)
			}
			if _, err := snap.RestoreStrict(); !errors.Is(err, ErrUnknownEffectKey) {
				t.Errorf("RestoreStrict: err = %v, want ErrUnknownEffectKey", err)
			}
		})
	}
	// The unaltered file is fine.
	var snap GameSnapshot
	if err := json.Unmarshal(raw, &snap); err != nil {
		t.Fatal(err)
	}
	if _, err := snap.RestoreStrict(); err != nil {
		t.Errorf("an intact file was refused: %v", err)
	}
}
