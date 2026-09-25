package game

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"
)

// scoped_grants_test.go — ADR 0093 PR 4 (#1584): the grantAbilities
// mod of ADR 0041 phase 3's ScopedEffect record. A resolving spell's
// "gains '<ability>' until end of turn" lands on the recipient exactly
// as a granting static's does (the same layer-6 declaration), ends on
// its duration, stays pinned to the object it began on (CR 611.2c,
// CR 400.7), sorts by timestamp against a removal (CR 613.6), and
// survives a restore — while a restore point naming a bundle this
// binary does not register is refused.

const (
	sgBundle      = "scoped-grant-fixture/tap-for-u"
	sgOtherBundle = "scoped-grant-fixture/tap-for-g"
)

// stubScopedGrantBundles registers two mana bundles as the catalog.
func stubScopedGrantBundles(t *testing.T) {
	t.Helper()
	defs := map[string]*CardDef{
		GrantKey(sgBundle): {
			ManaAbilities: []ManaAbilityShape{{TapCost: true, Produced: "{U}", Label: "Add {U}"}},
			GrantText:     "{T}: Add {U}.",
		},
		GrantKey(sgOtherBundle): {
			ManaAbilities: []ManaAbilityShape{{TapCost: true, Produced: "{G}", Label: "Add {G}"}},
			GrantText:     "{T}: Add {G}.",
		},
	}
	prev := CatalogLookup
	CatalogLookup = func(key string) *CardDef { return defs[key] }
	t.Cleanup(func() { CatalogLookup = prev })
}

// sgRegister registers one record on `id` with its source `src` and
// reports it registered.
func sgRegister(t *testing.T, g *Game, src, id uuid.UUID, mods []Mod, d func() Duration) {
	t.Helper()
	ok := false
	g.WithWriteLock(func() {
		ok = g.RegisterScopedEffectForEffect(src, g.PinnedObjectsLocked(id), mods, d(), "scoped grant test")
	})
	if !ok {
		t.Fatal("RegisterScopedEffectForEffect registered nothing")
	}
}

// sgGrants is the recipient's layer-6 grants after a recompute.
func sgGrants(t *testing.T, g *Game, id uuid.UUID) []GrantedAbility {
	t.Helper()
	return scopedEffectChar(t, g, id).GrantedAbilities
}

// sgGrantedManaRows counts the recipient's mana-ability rows that came
// from a grant — what the ability menu and the auto-tapper read.
func sgGrantedManaRows(t *testing.T, g *Game, id uuid.UUID) int {
	t.Helper()
	n := 0
	g.WithWriteLock(func() {
		g.RecomputeLayersIfStaleLocked()
		c, ok := g.battlefieldCardLocked(id)
		if !ok {
			t.Fatalf("card %s not on battlefield", id)
		}
		_, origins := ManaAbilitiesWithOrigins(*c)
		for _, o := range origins {
			if o.Granted() {
				n++
			}
		}
	})
	return n
}

func TestAGrantAbilitiesModGivesTheBundleInLayer6(t *testing.T) {
	stubScopedGrantBundles(t)
	g := newActiveGame(t)
	me := g.Seats[0].ID
	helix := pushScopedTestCreature(g, me, 1, 1) // stands in for the spell's source
	bear := pushScopedTestCreature(g, me, 2, 2)
	sgRegister(t, g, helix, bear, []Mod{GrantAbilitiesMod(sgBundle)}, func() Duration { return IndefiniteDuration() })

	grants := sgGrants(t, g, bear)
	if len(grants) != 1 || grants[0].Key != GrantKey(sgBundle) || grants[0].Source != helix {
		t.Fatalf("granted = %+v, want one %s from the record's source", grants, GrantKey(sgBundle))
	}
	if got := sgGrantedManaRows(t, g, bear); got != 1 {
		t.Errorf("granted mana rows = %d, want the bundle's one", got)
	}
	if other := sgGrants(t, g, helix); len(other) != 0 {
		t.Errorf("an object outside the pinned set got %+v", other)
	}
}

func TestGrantAbilitiesModStoresTheGrantKeyForm(t *testing.T) {
	m := GrantAbilitiesMod("a/one", "grant:a/two", "")
	if m.Kind != ModGrantAbilities || len(m.Grants) != 2 ||
		m.Grants[0] != "grant:a/one" || m.Grants[1] != "grant:a/two" {
		t.Errorf("mod = %+v, want two keys in GrantKey form and the empty one dropped", m)
	}
}

func TestRegisteringAGrantModWithNoBundlePanics(t *testing.T) {
	g := newActiveGame(t)
	id := pushScopedTestCreature(g, g.Seats[0].ID, 2, 2)
	defer func() {
		if recover() == nil {
			t.Error("a grantAbilities mod naming nothing registered without a panic")
		}
	}()
	g.WithWriteLock(func() {
		g.RegisterScopedEffectForEffect(uuid.Nil, g.PinnedObjectsLocked(id),
			[]Mod{GrantAbilitiesMod()}, IndefiniteDuration(), "test")
	})
}

// "Until end of turn": the grant is there through the turn and gone
// after the cleanup sweep (CR 514.2).
func TestAnUntilEndOfTurnGrantEndsAtCleanup(t *testing.T) {
	stubScopedGrantBundles(t)
	g := newActiveGame(t)
	bear := pushScopedTestCreature(g, g.Seats[0].ID, 2, 2)
	var eot Duration
	g.WithWriteLock(func() { eot = g.UntilEndOfTurnDuration() })
	sgRegister(t, g, uuid.Nil, bear, []Mod{GrantAbilitiesMod(sgBundle)}, func() Duration { return eot })
	if len(sgGrants(t, g, bear)) != 1 {
		t.Fatal("setup: the grant applies")
	}
	g.WithWriteLock(func() { g.ClearExpiredScopedStaticsLocked() })
	if len(sgGrants(t, g, bear)) != 1 {
		t.Fatal("an ordinary sweep mid-turn ended an until-end-of-turn grant")
	}
	g.WithWriteLock(func() { g.ClearEndOfTurnScopedStaticsLocked() })
	if got := sgGrants(t, g, bear); len(got) != 0 {
		t.Errorf("after cleanup the creature still has %+v", got)
	}
	if len(g.ScopedEffects) != 0 {
		t.Errorf("%d records outlived their duration", len(g.ScopedEffects))
	}
}

// CR 611.2c / CR 400.7: the set is locked on the object. A creature
// that leaves and comes back is a new object the grant never named.
func TestAGrantDoesNotFollowItsObjectThroughAFlicker(t *testing.T) {
	stubScopedGrantBundles(t)
	g := newActiveGame(t)
	me := g.Seats[0].ID
	bear := pushScopedTestCreature(g, me, 2, 2)
	sgRegister(t, g, uuid.Nil, bear, []Mod{GrantAbilitiesMod(sgBundle)}, func() Duration { return IndefiniteDuration() })
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == bear {
				// What a flicker leaves behind: the same instance with
				// a new entry stamp.
				g.Battlefield.Cards[i].EnteredBattlefieldAt += 1_000_000
			}
		}
		g.layerVersion.Add(1)
	})
	if got := sgGrants(t, g, bear); len(got) != 0 {
		t.Errorf("the new object has %+v; the grant was pinned to the old one", got)
	}
}

// CR 613.6: a removal applies in its own timestamp slot. A "loses all
// abilities" EARLIER than the grant cannot reach it; one LATER takes it.
func TestAGrantSortsAgainstARemovalByTimestamp(t *testing.T) {
	stubScopedGrantBundles(t)
	for _, tc := range []struct {
		name          string
		removalFirst  bool
		wantGrantKept bool
	}{
		{"removal then grant", true, true},
		{"grant then removal", false, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tick := int64(1_000_000)
			restore := SetClockForTest(func() int64 { tick += 1000; return tick })
			defer restore()
			g := newActiveGame(t)
			bear := pushScopedTestCreature(g, g.Seats[0].ID, 2, 2)
			ind := func() Duration { return IndefiniteDuration() }
			grant, removal := []Mod{GrantAbilitiesMod(sgBundle)}, []Mod{LoseAllAbilitiesMod()}
			if tc.removalFirst {
				sgRegister(t, g, uuid.Nil, bear, removal, ind)
				sgRegister(t, g, uuid.Nil, bear, grant, ind)
			} else {
				sgRegister(t, g, uuid.Nil, bear, grant, ind)
				sgRegister(t, g, uuid.Nil, bear, removal, ind)
			}
			kept := len(sgGrants(t, g, bear)) == 1
			if kept != tc.wantGrantKept {
				t.Errorf("grant kept = %v, want %v", kept, tc.wantGrantKept)
			}
		})
	}
}

// "Loses all abilities and has '…'" as ONE record: the removal is
// applied before the grant in the same slot (ADR 0046 §2), whatever
// order the two mods sort in.
func TestLoseAllAbilitiesAndAGrantInOneRecordKeepsTheGrant(t *testing.T) {
	stubScopedGrantBundles(t)
	g := newActiveGame(t)
	bear := pushScopedTestCreature(g, g.Seats[0].ID, 2, 2)
	sgRegister(t, g, uuid.Nil, bear, []Mod{LoseAllAbilitiesMod(), GrantAbilitiesMod(sgBundle)},
		func() Duration { return IndefiniteDuration() })
	c := scopedEffectChar(t, g, bear)
	if !c.AbilitiesRemoved || len(c.GrantedAbilities) != 1 {
		t.Errorf("removed %v, grants %+v: want the own abilities gone and the grant kept", c.AbilitiesRemoved, c.GrantedAbilities)
	}
}

// The payoff: a table holding a duration grant is a full restore point,
// and the restored creature still has the ability.
func TestADurationGrantIsARestorePoint(t *testing.T) {
	stubScopedGrantBundles(t)
	g := newActiveGame(t)
	bear := pushScopedTestCreature(g, g.Seats[0].ID, 2, 2)
	sgRegister(t, g, uuid.Nil, bear, []Mod{GrantAbilitiesMod(sgBundle, sgOtherBundle)},
		func() Duration { return IndefiniteDuration() })

	snap := g.CaptureSnapshot()
	if !snap.Restorable() {
		t.Fatalf("a game holding a duration grant is not a restore point: %+v", snap.Continuations)
	}
	raw, err := json.Marshal(snap)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw), `"grants":["grant:`+sgBundle+`","grant:`+sgOtherBundle+`"]`) {
		t.Errorf("the record's bundles are not on disk as data:\n%s", raw)
	}
	var decoded GameSnapshot
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatal(err)
	}
	restored, err := decoded.RestoreStrict()
	if err != nil {
		t.Fatalf("RestoreStrict: %v", err)
	}
	if got := sgGrants(t, restored, bear); len(got) != 2 {
		t.Errorf("restored grants = %+v, want both bundles", got)
	}
	if got := sgGrantedManaRows(t, restored, bear); got != 2 {
		t.Errorf("restored granted mana rows = %d, want 2", got)
	}
}

// ADR 0041 P4: a restore point naming a bundle this binary's catalog
// does not register — a newer build's card — is refused, not restored
// as a creature with a grant that silently does nothing. So is a mod
// naming no bundle, which no binary writes.
func TestAnUnknownGrantBundleIsRefused(t *testing.T) {
	stubScopedGrantBundles(t)
	for name, grants := range map[string][]string{
		"unregistered bundle": {GrantKey("scoped-grant-fixture/from-the-future")},
		"no bundle":           nil,
		"empty key":           {""},
	} {
		t.Run(name, func(t *testing.T) {
			g := newActiveGame(t)
			bear := pushScopedTestCreature(g, g.Seats[0].ID, 2, 2)
			sgRegister(t, g, uuid.Nil, bear, []Mod{GrantAbilitiesMod(sgBundle)}, func() Duration { return IndefiniteDuration() })
			snap := g.CaptureSnapshot()
			snap.ScopedEffects[0].Mods[0].Grants = grants
			if _, err := snap.Restore(); !errors.Is(err, ErrUnknownEffectKey) {
				t.Errorf("Restore: err = %v, want ErrUnknownEffectKey", err)
			}
			if _, err := snap.RestoreStrict(); !errors.Is(err, ErrUnknownEffectKey) {
				t.Errorf("RestoreStrict: err = %v, want ErrUnknownEffectKey", err)
			}
		})
	}
}

// CR 707.2: a layer-6 grant is not a copiable value, so a copy of the
// recipient does not have it — structurally, because the copiable
// values never read the layered characteristic.
func TestACopyOfARecipientDoesNotCopyTheDurationGrant(t *testing.T) {
	stubScopedGrantBundles(t)
	g := newActiveGame(t)
	bear := pushScopedTestCreature(g, g.Seats[0].ID, 2, 2)
	sgRegister(t, g, uuid.Nil, bear, []Mod{GrantAbilitiesMod(sgBundle)}, func() Duration { return IndefiniteDuration() })
	var copied PrintedValues
	g.WithWriteLock(func() {
		g.RecomputeLayersIfStaleLocked()
		c, _ := g.battlefieldCardLocked(bear)
		copied = CopiableValuesOf(*c)
	})
	if len(copied.GrantedAbilities) != 0 {
		t.Errorf("the copiable values carry %v", copied.GrantedAbilities)
	}
}
