package game

import (
	"encoding/json"
	"reflect"
	"testing"
)

// copy_grants_test.go — CR 707.9a / CR 707.9b, the engine half.
// What a granted ability has to survive: being copied AGAIN, an
// undo, a snapshot, ability removal and a face-down flip.

// stubGrantCatalog wires a catalog holding one card ("oracle-bear",
// with an ETB trigger of its own) and one ability grant, and returns
// the grant's bare name.
func stubGrantCatalog(t *testing.T) string {
	t.Helper()
	prev := CatalogLookup
	t.Cleanup(func() { CatalogLookup = prev })
	defs := map[string]*CardDef{
		"oracle-bear": {
			Triggered: []TriggeredAbility{{Watches: []EventKind{EventETB}}},
			Static:    []StaticAbility{{Layer: Layer6Ability}},
		},
		GrantKey("probe/illusion"): {
			Triggered: []TriggeredAbility{{Watches: []EventKind{EventBecomesTarget}}},
			Activated: []ActivatedAbilityShape{{Label: "{2}: do the thing"}},
		},
	}
	CatalogLookup = func(key string) *CardDef { return defs[key] }
	return "probe/illusion"
}

// TestAddSubtypeAddsToTheTypeLine — CR 707.9b, "it's an Illusion in
// ADDITION to its other types". The card types and supertypes are
// untouched and the printed subtypes stay.
func TestAddSubtypeAddsToTheTypeLine(t *testing.T) {
	v := PrintedValues{TypeLine: "Legendary Creature — Human Wizard"}
	v.AddSubtype("Illusion")
	if v.TypeLine != "Legendary Creature — Human Wizard Illusion" {
		t.Fatalf("type line = %q", v.TypeLine)
	}
	v.AddSubtype("Illusion")
	if v.TypeLine != "Legendary Creature — Human Wizard Illusion" {
		t.Errorf("AddSubtype is not idempotent: %q", v.TypeLine)
	}
	if !v.HasSubtype("Illusion") || !v.HasSubtype("Wizard") || v.HasSubtype("Bear") {
		t.Error("HasSubtype disagrees with the type line it just wrote")
	}
	v.RemoveSubtype("Human")
	if v.TypeLine != "Legendary Creature — Wizard Illusion" {
		t.Errorf("after RemoveSubtype, type line = %q", v.TypeLine)
	}
	v.RemoveSubtype("Human")
	if v.TypeLine != "Legendary Creature — Wizard Illusion" {
		t.Errorf("RemoveSubtype of an absent subtype changed the line: %q", v.TypeLine)
	}
}

// TestAddSubtypeOnATypeLineWithNoSubtypes covers the em-dash the
// composer has to invent — a copy of Ornithopter that becomes an
// Illusion needs the separator that was never printed.
func TestAddSubtypeOnATypeLineWithNoSubtypes(t *testing.T) {
	v := PrintedValues{TypeLine: "Artifact Creature"}
	v.AddSubtype("Illusion")
	if v.TypeLine != "Artifact Creature — Illusion" {
		t.Fatalf("type line = %q, want the em-dash added", v.TypeLine)
	}
	if _, _, subs := ParseTypeLine(v.TypeLine); len(subs) != 1 || subs[0] != "Illusion" {
		t.Errorf("the line does not parse back: %v", subs)
	}
}

// TestAddSubtypeRewritesTheActiveFace keeps ADR 0034's invariant —
// the flat printed fields equal Faces[ActiveFace] — true after an
// except clause edits the type line.
func TestAddSubtypeRewritesTheActiveFace(t *testing.T) {
	v := PrintedValues{
		TypeLine:   "Creature — Bear",
		Faces:      []Face{{TypeLine: "Creature — Bear"}, {TypeLine: "Land"}},
		ActiveFace: 0,
	}
	v.AddSubtype("Illusion")
	if v.Faces[0].TypeLine != v.TypeLine {
		t.Errorf("face 0 type line = %q, flat = %q — they must agree", v.Faces[0].TypeLine, v.TypeLine)
	}
	if v.Faces[1].TypeLine != "Land" {
		t.Errorf("the inactive face was rewritten: %q", v.Faces[1].TypeLine)
	}
}

// TestGrantAbilityIsIdempotentAndKeyed — the copy stores catalog
// keys, one per bundle, in the order the clause granted them.
func TestGrantAbilityIsIdempotentAndKeyed(t *testing.T) {
	var v PrintedValues
	v.GrantAbility("a/one")
	v.GrantAbility("a/one")
	v.GrantAbility(GrantKey("a/two"))
	v.GrantAbility("")
	want := []string{"grant:a/one", "grant:a/two"}
	if !reflect.DeepEqual(v.GrantedAbilities, want) {
		t.Fatalf("granted = %v, want %v", v.GrantedAbilities, want)
	}
}

// TestCatalogKeyCarriesGrants pins the seam: an object with no
// grants keys exactly as it always did, and one with grants keys to
// a composite the base of which is still recoverable.
func TestCatalogKeyCarriesGrants(t *testing.T) {
	plain := Card{OracleID: "oracle-bear"}
	if got := CatalogKey(plain); got != "oracle-bear" {
		t.Fatalf("a card with no grants must key unchanged, got %q", got)
	}
	granted := Card{OracleID: "oracle-bear", GrantedAbilities: []string{"grant:x", "grant:y"}}
	key := CatalogKey(granted)
	if key != "oracle-bear|grant:x|grant:y" {
		t.Fatalf("composite key = %q", key)
	}
	if got := BaseCatalogKey(key); got != "oracle-bear" {
		t.Errorf("BaseCatalogKey = %q, want the card identity back", got)
	}
	if got := BaseCatalogKey("oracle-bear#1"); got != "oracle-bear#1" {
		t.Errorf("BaseCatalogKey must leave a face key alone, got %q", got)
	}
	back := Card{OracleID: "oracle-bear", ActiveFace: 1, GrantedAbilities: []string{"grant:x"}}
	if got := CatalogKey(back); got != "oracle-bear#1|grant:x" {
		t.Errorf("a back face with a grant keys %q", got)
	}
}

// TestFaceDownSuppressesGrantedAbilities — CR 708.2a. A face-down
// permanent has NO text, and a granted ability is text.
func TestFaceDownSuppressesGrantedAbilities(t *testing.T) {
	c := Card{
		OracleID:         "oracle-bear",
		GrantedAbilities: []string{"grant:x"},
		FaceDown:         true,
		FaceDownKind:     FaceDownManifested,
	}
	if !c.FaceDownIsPermanent() {
		t.Skip("the face-down kind used here is not a permanent one")
	}
	if got := CatalogKey(c); got != "" {
		t.Errorf("CatalogKey = %q, want the CR 708.2a silence to win over the grant", got)
	}
}

// TestGrantedAbilitiesAreFoundThroughTheOneLookup is the whole
// point: every reader in the engine goes through CatalogKey +
// catalogDef, so a granted trigger, static and activated ability all
// turn up with no reader of their own.
func TestGrantedAbilitiesAreFoundThroughTheOneLookup(t *testing.T) {
	grant := stubGrantCatalog(t)

	c := Card{OracleID: "oracle-bear"}
	c.GrantedAbilities = []string{GrantKey(grant)}
	key := CatalogKey(c)

	triggers := CatalogTriggers(key)
	if len(triggers) != 2 {
		t.Fatalf("triggers = %d, want the card's ETB plus the granted one", len(triggers))
	}
	if triggers[0].Watches[0] != EventETB || triggers[1].Watches[0] != EventBecomesTarget {
		t.Error("the card's own abilities must come first, the granted ones after")
	}
	if got := CatalogActivatedAbilities(key); len(got) != 1 || got[0].Label != "{2}: do the thing" {
		t.Errorf("granted activated abilities = %v", got)
	}
	if got := CatalogStaticAbilities(key); len(got) != 1 {
		t.Errorf("statics = %d, want the card's own kept", len(got))
	}
	if !IsCatalogCard(key) {
		t.Error("a granted copy must still read as a catalog object")
	}
}

// TestAGrantSurvivesOnACardWithNoCatalogEntry — the normal case.
// Phantasmal Image copies an imported vanilla bear: the bear has no
// spec, and the copy still has the sacrifice trigger.
func TestAGrantSurvivesOnACardWithNoCatalogEntry(t *testing.T) {
	grant := stubGrantCatalog(t)

	c := Card{OracleID: "oracle-nothing", GrantedAbilities: []string{GrantKey(grant)}}
	key := CatalogKey(c)
	if got := CatalogTriggers(key); len(got) != 1 || got[0].Watches[0] != EventBecomesTarget {
		t.Fatalf("triggers = %v, want just the granted one", got)
	}
	if !IsCatalogCard(key) {
		t.Error("an object whose only abilities are granted still has abilities")
	}
	// And a grant nobody registered is not an object at all.
	missing := Card{OracleID: "oracle-nothing", GrantedAbilities: []string{GrantKey("no/such")}}
	if IsCatalogCard(CatalogKey(missing)) {
		t.Error("an unregistered grant must not conjure a catalog entry")
	}
}

// TestMergingAGrantDoesNotMutateTheCatalog is the long-fuse bug the
// concat helpers exist to prevent: the base *CardDef is shared by
// every card with that oracle ID.
func TestMergingAGrantDoesNotMutateTheCatalog(t *testing.T) {
	grant := stubGrantCatalog(t)

	before := len(CatalogTriggers("oracle-bear"))
	c := Card{OracleID: "oracle-bear", GrantedAbilities: []string{GrantKey(grant)}}
	_ = CatalogTriggers(CatalogKey(c))
	if after := len(CatalogTriggers("oracle-bear")); after != before {
		t.Fatalf("the card's own trigger list grew from %d to %d — the merge wrote into the catalog", before, after)
	}
}

// TestAbilityRemovalTakesGrantedAbilitiesToo — CR 613.1f. A granted
// ability is an ability of the object like any other.
func TestAbilityRemovalTakesGrantedAbilitiesToo(t *testing.T) {
	grant := stubGrantCatalog(t)

	c := Card{OracleID: "oracle-bear", GrantedAbilities: []string{GrantKey(grant)}}
	c.effective = &Characteristic{AbilitiesRemoved: true}
	if got := CatalogAbilityKey(c); got != "" {
		t.Errorf("CatalogAbilityKey = %q, want the removal to silence the grant too", got)
	}
}

// TestAGrantIsACopiableValue is CR 707.9a's second sentence: a Clone
// copying a Phantasmal Image gets the Illusion type AND the trigger,
// because both ride in the copiable values.
func TestAGrantIsACopiableValue(t *testing.T) {
	image := Card{
		OracleID: "oracle-bear",
		Name:     "Grizzly Bears",
		TypeLine: "Creature — Bear Illusion",
	}
	image.GrantedAbilities = []string{GrantKey("probe/illusion")}

	values := CopiableValuesOf(image)
	if !reflect.DeepEqual(values.GrantedAbilities, image.GrantedAbilities) {
		t.Fatalf("copiable values dropped the grant: %v", values.GrantedAbilities)
	}

	clone := Card{OracleID: "oracle-clone", Name: "Clone", TypeLine: "Creature — Shapeshifter"}
	clone.applyCopy(values, image)
	if !clone.HasSubtypeName("Illusion") {
		t.Errorf("the clone's type line is %q, want the added subtype copied", clone.TypeLine)
	}
	if !reflect.DeepEqual(clone.GrantedAbilities, image.GrantedAbilities) {
		t.Fatalf("the clone did not inherit the grant: %v", clone.GrantedAbilities)
	}
	// And it reverts: the card in the graveyard is a Clone again.
	clone.restorePrintedSelf()
	if len(clone.GrantedAbilities) != 0 || clone.TypeLine != "Creature — Shapeshifter" {
		t.Errorf("after leaving the battlefield: type %q, grants %v", clone.TypeLine, clone.GrantedAbilities)
	}
}

// HasSubtypeName is a test-local spelling of "does this printed type
// line carry this subtype", so the assertion above does not depend on
// the layer engine having run.
func (c Card) HasSubtypeName(s string) bool {
	_, _, subs := ParseTypeLine(c.TypeLine)
	for _, x := range subs {
		if x == s {
			return true
		}
	}
	return false
}

// TestEveryCopiableValueSurvivesApplyCopy is the drift guard
// PrintedValues never had: a field added to the copiable values and
// forgotten in applyCopy or CopiableValuesOf is silently dropped by
// every copy in the game, and no existing test would say so.
func TestEveryCopiableValueSurvivesApplyCopy(t *testing.T) {
	src := PrintedValues{}
	rv := reflect.ValueOf(&src).Elem()
	rt := rv.Type()
	for i := 0; i < rt.NumField(); i++ {
		f := rv.Field(i)
		switch f.Kind() {
		case reflect.String:
			f.SetString("x" + rt.Field(i).Name)
		case reflect.Int:
			f.SetInt(7)
		case reflect.Bool:
			f.SetBool(true)
		case reflect.Slice:
			if rt.Field(i).Type.Elem().Kind() == reflect.String {
				f.Set(reflect.ValueOf([]string{"a", "b"}))
			} else {
				f.Set(reflect.MakeSlice(rt.Field(i).Type, 1, 1))
			}
		}
	}
	// ActiveFace has to index into Faces for the multi-face
	// invariant to mean anything; every other field is filled above.
	src.ActiveFace = 0

	var dst Card
	dst.applyCopy(src, Card{})
	got := CopiableValuesOf(dst)
	if !reflect.DeepEqual(got, src) {
		t.Fatalf(`a copiable value did not survive applyCopy → CopiableValuesOf.

got  %+v
want %+v

A field was added to PrintedValues and one of the five hand-written
copy sites was missed. They are CopiableValuesOf, PrintedValues.Clone,
Card.applyCopy and Card.restorePrintedSelf in copy.go, plus
copyPrintedValues in snapshot.go.`, got, src)
	}
}

// TestGrantedAbilitiesSurviveASnapshotRoundTrip — the grant is
// carried, and it is carriable at all because it holds names rather
// than closures.
func TestGrantedAbilitiesSurviveASnapshotRoundTrip(t *testing.T) {
	g := newActiveGame(t)
	owner := g.Seats[0]
	srcID := bearsOnBattlefield(g, owner.ID, "Grizzly Bears")

	g.mu.Lock()
	src, _ := g.battlefieldCardLocked(srcID)
	values := CopiableValuesOf(*src)
	values.AddSubtype("Illusion")
	values.GrantAbility("probe/illusion")
	src.applyCopy(values, *src)
	g.mu.Unlock()

	snap := g.CaptureSnapshot()
	blob, err := json.Marshal(snap)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var round GameSnapshot
	if err := json.Unmarshal(blob, &round); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	restored, err := round.Restore()
	if err != nil {
		t.Fatalf("restore: %v", err)
	}
	var found *Card
	for i := range restored.Battlefield.Cards {
		if restored.Battlefield.Cards[i].InstanceID == srcID {
			found = &restored.Battlefield.Cards[i]
		}
	}
	if found == nil {
		t.Fatal("the copied permanent is not on the restored battlefield")
	}
	if !reflect.DeepEqual(found.GrantedAbilities, []string{"grant:probe/illusion"}) {
		t.Fatalf("restored grants = %v", found.GrantedAbilities)
	}
	if found.PrintedSelf == nil || len(found.PrintedSelf.GrantedAbilities) != 0 {
		t.Error("PrintedSelf must come back carrying the card's OWN (empty) grant list")
	}
	if CatalogKey(*found) != "oracle-Grizzly Bears|grant:probe/illusion" {
		t.Errorf("the restored permanent keys as %q", CatalogKey(*found))
	}
}

// TestUndoAcrossACopyRestoresTheGrantIndependently — clone.go deep
// copies the slice, so replaying the copy cannot reach back into the
// undo snapshot's.
func TestUndoAcrossACopyRestoresTheGrantIndependently(t *testing.T) {
	g := newActiveGame(t)
	owner := g.Seats[0]
	srcID := bearsOnBattlefield(g, owner.ID, "Grizzly Bears")

	g.mu.Lock()
	src, _ := g.battlefieldCardLocked(srcID)
	before := cloneCard(*src)
	values := CopiableValuesOf(*src)
	values.GrantAbility("probe/illusion")
	src.applyCopy(values, *src)
	g.mu.Unlock()

	if len(before.GrantedAbilities) != 0 {
		t.Fatalf("the pre-copy clone grew a grant: %v", before.GrantedAbilities)
	}
	g.mu.Lock()
	after, _ := g.battlefieldCardLocked(srcID)
	afterClone := cloneCard(*after)
	afterClone.GrantedAbilities[0] = "grant:tampered"
	stillLive := after.GrantedAbilities[0]
	g.mu.Unlock()
	if stillLive != "grant:probe/illusion" {
		t.Errorf("mutating the clone reached the live card: %q", stillLive)
	}
}
